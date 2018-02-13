## Contributing
We always welcome pull requests. This guide will help you to ensure that that your time is spent wisely. 

If you are thinking of introduing a new feature or making a significant change to an existing behavior, discuss it
first with the maintainers by creating an issue.

### Process
- Fork and clone the repo
- Use a text editor with support for .editorconfig
- Implement the change
- Ensure a new test cases are added to cover new code
- Ensure you can run `make build` command without errors
- Commit and push your changes to your private fork. If you have multiple commits, squash them into one.
- Submit a PR

### Structure
mbt source is organised into a few go modules that resides in the repository root.

|Module |Description|
|-----|------|
|lib |This is the main implementation of mbt functionality |
|cmd |Cobra commands for CLI |
|dtrace |Utility for writing debug trace messages from other packages |
|e |go error implementation enriched with callsite info |
|fsutil |File IO utility |
|trie |Prefix trie implementation |

### Writing Code
- When you return an error from received from call to an external library, wrap it in an `E` using `e.Wrap` function. Use 
the other `e.Xxx` methods if you want to enrich the error with more information.


