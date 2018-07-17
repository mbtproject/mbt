## 0.21.0
- [6c9f0d4](https://github.com/mbtproject/mbt/commit/6c9f0d4) [feat] Fail fast option for user defined commands
- [f53f196](https://github.com/mbtproject/mbt/commit/f53f196) [a4ab9ae](https://github.com/mbtproject/mbt/commit/a4ab9ae) [a9903c4](https://github.com/mbtproject/mbt/commit/a9903c4) [feat] Default build command

## 0.20.1
- [e3a0fbb](https://github.com/mbtproject/mbt/commit/e3a0fbb) [f855765](https://github.com/mbtproject/mbt/commit/f855765) [perf] Faster discovery of local modules

## 0.20.0

- [398b10e](https://github.com/mbtproject/mbt/commit/398b10e) [2af9407](https://github.com/mbtproject/mbt/commit/2af9407) Feature: Additional environment values

## 0.19.0

- [529068f](https://github.com/mbtproject/mbt/commit/529068f) Feature: Make a sorted modules list available for templates

## 0.18.0

- [9ed7c00](https://github.com/mbtproject/mbt/commit/9ed7c00) Feature: New template helpers

## 0.17.1

- [24e3039](https://github.com/mbtproject/mbt/commit/24e3039) Fix: Support builds when head is detached

## 0.17.0

- [97855c3](https://github.com/mbtproject/mbt/commit/97855c3) Fix: Version is written to stdout instead of stderr
- [aecdcd3](https://github.com/mbtproject/mbt/commit/aecdcd3) Feature: Template helper to convert a map to a list
- [d8b898a](https://github.com/mbtproject/mbt/commit/d8b898a) Fix: Prevent build|run-in commands when head is detached
- [ff5d22a](https://github.com/mbtproject/mbt/commit/ff5d22a) Feature: Execute user defined commands in modules

## 0.16.0
### Breaking Changes
This version includes significant breaking changes. Click through to the 
commit messages for more information.

- [cdcc122](https://github.com/mbtproject/mbt/commit/cdcc122) Fix: Consider file dependencies when calculating the version
- [c52b91b](https://github.com/mbtproject/mbt/commit/c52b91b) Feature: Filter modules during commands based on git tree

## 0.15.1
- [3f20eee](https://github.com/mbtproject/mbt/commit/3f20eee) Fix: Update head during checkout operations
- [0760617](https://github.com/mbtproject/mbt/commit/0760617) Fix: Detect local changes on root
- [60792ab](https://github.com/mbtproject/mbt/commit/60792ab) Fix: Handle root module changes correctly

## 0.15.0
- [b02aaad](https://github.com/mbtproject/mbt/commit/b02aaad) Feature: Filter modules during local build/describe
- [7c0f5b0](https://github.com/mbtproject/mbt/commit/7c0f5b0) Feature: Run mbt command from anywhere
- [78c1b44](https://github.com/mbtproject/mbt/commit/78c1b44) Feature: Improved cyclic error message
- [25e3ae6](https://github.com/mbtproject/mbt/commit/25e3ae6) Feature: Display a build summary
- [884819e](https://github.com/mbtproject/mbt/commit/884819e) Feature: Display an empty list when no modules to describe.
- [b0796b0](https://github.com/mbtproject/mbt/commit/b0796b0) Feature: Build/Describe content of a commit
