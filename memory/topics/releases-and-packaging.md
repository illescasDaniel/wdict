# Releases and packaging
Updated: 2026-11-16. Related: [pronunciation](pronunciation.md)

## Now
- Releases come only from the dispatched workflow ([release-via-dispatched-workflow](../notes/release-via-dispatched-workflow.md)); after publishing, update the user's install ([update-system-install-after-release](../notes/update-system-install-after-release.md)). Versions: [version-reporting-ldflags](../notes/version-reporting-ldflags.md).
- CI fails npm audit at moderate ([npm-audit-fails-on-moderate](../notes/npm-audit-fails-on-moderate.md)); dependencies stay project-local ([dependencies-project-local](../notes/dependencies-project-local.md)).
- 0.8.4 is pushed but unreleased while the picomatch advisory has no fix ([picomatch-advisory-blocks-0-8-4](../notes/picomatch-advisory-blocks-0-8-4.md)).
- AppImage: [appimage-wails3-generator](../notes/appimage-wails3-generator.md), eSpeak bundled ([espeak-bundled-in-appimage](../notes/espeak-bundled-in-appimage.md)). Attribution rule: [attribution-pragmatic-review](../notes/attribution-pragmatic-review.md).

## How it got here
- First release was to wait for macOS and Windows ([first-release-deferred-until-all-platforms](../notes/first-release-deferred-until-all-platforms.md)); v0.1.0 shipped Linux-only on 2026-10-28 ([v0-1-0-linux-first-release](../notes/v0-1-0-linux-first-release.md)).
- Last shipped before this conversion: v0.8.3 (2026-11-15, native-player change confirmed by the user listening).
