# The `notwork` Analyzer

This module provides static analysis of code bases using
`github.com/thediveo/notwork`, including recommendations for fixes:
- refactoring the use of deprecated functions from the packages `netns` and
  `mntns`.
- fixing incorrect `defer EnterTransient()` usage.
 