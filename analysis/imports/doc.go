/*
Package imports supports analyzing and refactoring imports.

# Collecting Package Identifiers

[CollectPackageUsages] provides the “local name” and thus identifier assigned to
particular imports. This information is retrieved from actual package identifier
usage within a specific source file, but never guessed from the base names of
import paths.

Please note that dot and underscore imports are never included in the
information returned by CollectPackageUsages.
*/
package imports
