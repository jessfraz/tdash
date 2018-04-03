# Change Log

**ATTN**: This project uses [semantic versioning](http://semver.org/).

## [Unreleased]

## 0.2.0 - 2018-03-10

Backwards incomptabile changes:

* Switched `FeatureFlags` to a `struct` from `map[string]bool` as it was found that not all feature flags are `bool`s
  (https://github.com/jszwedko/go-circleci/issues/8) which resulted in non-bool values being inaccessible. Known feature
  flags are encoded as struct fields with a `.Raw()` method to access the underlying `map[string]interface{}` to access
  unknown feature flags.

## 0.1.0 - 2018-03-10

### Added
- Initial implementation.
