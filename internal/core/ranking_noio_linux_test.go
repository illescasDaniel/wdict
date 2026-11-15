//go:build linux && amd64

package core

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

// installRankingIOTrap makes accidental filesystem, networking, or subprocess
// access fatal on this locked thread, including calls that ignore their errors.
func installRankingIOTrap() error {
	calls := []uint32{unix.SYS_OPEN, unix.SYS_OPENAT, unix.SYS_OPENAT2, unix.SYS_CREAT, unix.SYS_SOCKET, unix.SYS_SOCKETPAIR, unix.SYS_CONNECT, unix.SYS_EXECVE, unix.SYS_EXECVEAT, unix.SYS_STAT, unix.SYS_LSTAT, unix.SYS_NEWFSTATAT, unix.SYS_READLINK, unix.SYS_READLINKAT}
	code := []unix.SockFilter{{Code: unix.BPF_LD | unix.BPF_W | unix.BPF_ABS, K: 0}}
	for _, nr := range calls {
		code = append(code, unix.SockFilter{Code: unix.BPF_JMP | unix.BPF_JEQ | unix.BPF_K, Jt: 0, Jf: 1, K: nr}, unix.SockFilter{Code: unix.BPF_RET | unix.BPF_K, K: unix.SECCOMP_RET_TRAP})
	}
	code = append(code, unix.SockFilter{Code: unix.BPF_RET | unix.BPF_K, K: unix.SECCOMP_RET_ALLOW})
	prog := unix.SockFprog{Len: uint16(len(code)), Filter: &code[0]}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return err
	}
	return unix.Prctl(unix.PR_SET_SECCOMP, unix.SECCOMP_MODE_FILTER, uintptr(unsafe.Pointer(&prog)), 0, 0)
}

func TestGivenRankingCandidatesWhenOrderedThenNoFilesystemNetworkOrProcessIO(t *testing.T) {
	// given: isolated process prevents the filter affecting other tests.
	if mode := os.Getenv("WDICT_RANKING_NOIO_CHILD"); mode != "" {
		runtime.LockOSThread()
		if err := installRankingIOTrap(); err != nil {
			t.Fatalf("install syscall filter: %v", err)
		}
		if mode == "probe" {
			unix.Open("/dev/null", unix.O_RDONLY, 0)
			t.Fatal("syscall trap failed")
		}
		// when: exercise exact, casefolded, folded, indirect and POS comparisons.
		got := RankEntries("CAR", []Entry{{Headword: "Car", Meanings: []Meaning{{PartOfSpeech: "Proper noun"}}}, {Headword: "car", Meanings: []Meaning{{PartOfSpeech: "noun"}}}, {Headword: "-car"}, {Headword: "other"}})
		// then: result is correct without any prohibited I/O syscall.
		if len(got) != 4 || got[0].Headword != "car" {
			t.Fatalf("unexpected order %+v", got)
		}
		return
	}
	run := func(mode string) error {
		cmd := exec.Command(os.Args[0], "-test.run=^TestGivenRankingCandidatesWhenOrderedThenNoFilesystemNetworkOrProcessIO$")
		cmd.Env = append(os.Environ(), "WDICT_RANKING_NOIO_CHILD="+mode)
		return cmd.Run()
	}
	if err := run("probe"); err == nil {
		t.Fatal("negative control unexpectedly allowed filesystem access")
	}
	if err := run("ranking"); err != nil {
		t.Fatalf("ranking attempted prohibited I/O or filter failed: %v", err)
	}
}
