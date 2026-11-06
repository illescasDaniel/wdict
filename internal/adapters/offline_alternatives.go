package adapters

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ianlewis/go-stardict/dict"
	"github.com/illescasDaniel/wdict/internal/core"
)

const (
	alternativesIndexReadLimit   = 128 << 20
	alternativesMaxRecordBytes   = 4096
	alternativesMaxMetadataBytes = 64 << 10
)

// StarDictHeadwordIndex opens independent index handles for each scan, so a
// suggestion scan never holds OfflineDictionary's exact-lookup lock.
type StarDictHeadwordIndex struct {
	IFO      string
	Language core.Language
}

// IterateHeadwords streams original index and synonym words. A successful
// return means both files were read completely and all synonym targets exist.
func (i *StarDictHeadwordIndex) IterateHeadwords(ctx context.Context, language core.Language, visit func(string) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if language != i.Language || (language != core.English && language != core.Spanish) {
		return errors.New("dictionary language mismatch")
	}
	if visit == nil {
		return errors.New("nil headword visitor")
	}
	if filepath.Ext(i.IFO) != ".ifo" {
		return errors.New("invalid dictionary metadata path")
	}

	meta, err := os.Open(i.IFO)
	if err != nil {
		return err
	}
	metadata, err := io.ReadAll(io.LimitReader(meta, alternativesMaxMetadataBytes+1))
	closeErr := meta.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if len(metadata) > alternativesMaxMetadataBytes {
		return errors.New("dictionary metadata too large")
	}
	if !utf8.Valid(metadata) {
		return errors.New("invalid dictionary metadata encoding")
	}
	var idx64 bool
	var synCount, wordCount uint64
	var hasWordCount bool
	for _, line := range strings.Split(string(metadata), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "idxoffsetbits=") {
			switch strings.TrimPrefix(line, "idxoffsetbits=") {
			case "32":
				idx64 = false
			case "64":
				idx64 = true
			default:
				return errors.New("unsupported dictionary index offset width")
			}
		}
		if strings.HasPrefix(line, "wordcount=") {
			if hasWordCount {
				return errors.New("duplicate dictionary word count")
			}
			wordCount, err = strconv.ParseUint(strings.TrimPrefix(line, "wordcount="), 10, 32)
			if err != nil {
				return fmt.Errorf("invalid dictionary word count: %w", err)
			}
			hasWordCount = true
		}
		if strings.HasPrefix(line, "synwordcount=") {
			synCount, err = strconv.ParseUint(strings.TrimPrefix(line, "synwordcount="), 10, 32)
			if err != nil {
				return fmt.Errorf("invalid synonym count: %w", err)
			}
		}
	}

	if !hasWordCount {
		return errors.New("missing dictionary word count")
	}

	// Match the normal dictionary reader, including dictzip validation. A usable
	// index alone is insufficient when the corresponding definitions are absent.
	definitions, err := dict.NewFromIfoPath(i.IFO, nil)
	if err != nil {
		return fmt.Errorf("open dictionary definitions: %w", err)
	}
	if err := definitions.Close(); err != nil {
		return fmt.Errorf("close dictionary definitions: %w", err)
	}

	base := strings.TrimSuffix(i.IFO, ".ifo")
	index, err := os.Open(base + ".idx")
	if err != nil {
		return err
	}
	defer index.Close()
	indexInfo, err := index.Stat()
	if err != nil {
		return err
	}
	if !indexInfo.Mode().IsRegular() || indexInfo.Size() < 0 || indexInfo.Size() > alternativesIndexReadLimit {
		return errors.New("invalid or oversized dictionary index")
	}
	syn, err := os.Open(base + ".syn")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) || synCount != 0 {
			return err
		}
		syn = nil
	}
	if syn != nil {
		defer syn.Close()
		info, err := syn.Stat()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > alternativesIndexReadLimit-indexInfo.Size() {
			return errors.New("invalid or oversized synonym index")
		}
	}

	budget := int64(alternativesIndexReadLimit)
	fixedIndexBytes := 8
	if idx64 {
		fixedIndexBytes = 12
	}
	count, err := scanAlternativeIndex(ctx, index, fixedIndexBytes, &budget, func(word string, _ []byte) error { return visit(word) })
	if err != nil {
		return err
	}
	if count != wordCount {
		return fmt.Errorf("dictionary word count mismatch: got %d, expected %d", count, wordCount)
	}
	if syn == nil {
		return nil
	}
	actualSynCount, err := scanAlternativeIndex(ctx, syn, 4, &budget, func(word string, data []byte) error {
		if uint64(binary.BigEndian.Uint32(data)) >= count {
			return errors.New("synonym target outside dictionary index")
		}
		return visit(word)
	})
	if err != nil {
		return err
	}
	if actualSynCount != synCount {
		return fmt.Errorf("synonym count mismatch: got %d, expected %d", actualSynCount, synCount)
	}
	return nil
}

// VerifyHeadword checks the entire index, so a later corrupt record is not
// mistaken for a complete successful scan.
func (i *StarDictHeadwordIndex) VerifyHeadword(ctx context.Context, candidate string, language core.Language) (string, bool, error) {
	var match string
	err := i.IterateHeadwords(ctx, language, func(word string) error {
		if match == "" && word == candidate {
			match = word
		}
		return nil
	})
	if err != nil {
		return "", false, err
	}
	return match, match != "", nil
}

// scanAlternativeIndex parses NUL-terminated UTF-8 names and fixed-size binary
// payloads with bounded records. ReadSlice is backed by a fixed-size buffer.
func scanAlternativeIndex(ctx context.Context, source io.Reader, fixedBytes int, budget *int64, visit func(string, []byte) error) (uint64, error) {
	reader := bufio.NewReaderSize(source, alternativesMaxRecordBytes+1)
	var count uint64
	var payload [12]byte
	for {
		if err := ctx.Err(); err != nil {
			return count, err
		}
		name, err := reader.ReadSlice(0)
		*budget -= int64(len(name))
		if *budget < 0 {
			return count, errors.New("dictionary index read budget exceeded")
		}
		if errors.Is(err, io.EOF) && len(name) == 0 {
			return count, nil
		}
		if err != nil {
			return count, fmt.Errorf("invalid dictionary index record: %w", err)
		}
		if len(name) <= 1 || len(name)-1 > alternativesMaxRecordBytes || !utf8.Valid(name[:len(name)-1]) {
			return count, errors.New("invalid dictionary headword")
		}
		// ReadSlice's result aliases the buffer, which ReadFull may refill.
		word := string(name[:len(name)-1])
		n, err := io.ReadFull(reader, payload[:fixedBytes])
		*budget -= int64(n)
		if *budget < 0 {
			return count, errors.New("dictionary index read budget exceeded")
		}
		if err != nil {
			return count, fmt.Errorf("truncated dictionary index record: %w", err)
		}
		if err := visit(word, payload[:fixedBytes]); err != nil {
			return count, err
		}
		count++
	}
}
