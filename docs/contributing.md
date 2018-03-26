# Contributing

The specification and code is licensed under the Apache 2.0 license found in
the [LICENSE](../LICENSE) file of this repository.

## Style

See the [Style Guide](../STYLEGUIDE.md).

## Sign your work

The sign-off is a simple line at the end of the explanation for the
patch, which certifies that you wrote it or otherwise have the right to
pass it on as an open-source patch.  The rules are pretty simple: if you
can certify the below (from
[developercertificate.org](http://developercertificate.org/)):

You can add the sign off when creating the git commit via:

```
git commit -s
```

then you just add a line to every git commit message:

```
Signed-off-by: Joe Smith <joe@gmail.com>
```

using your real name (sorry, no pseudonyms or anonymous contributions.)

## Pull Requests

* All code provided to openstorage should have unit tests.
* Large PRs should have first been proposed through as an issue. The PR should point back to the issue as `Implementation of #...`
* Try to keep PRs "small" to make it easy to review.

## API Change

### Adding a new API

* Create an issue and use the following template:

```

## API Request
NEW

## Why this is needed


## APIs (for each use the following)

**API Name:**
* **Description**: ...
* **Method**: GET / DELETE / ...
* **Endpoints**: /somePath
* **Params**:
  `<< Set of parameters expected over JSON or as part of the Path >>`
* **Result**:
 `<< JSON Output >>`
```

Once the APIs have been approved, then create a PR as the implementation of the issue. All APIs in the PR must have swagger documentation.

### Removing an API

To deprecate the API, we must first label the API as **deprecated** in swagger documentation and on the Release notes. After one release, the API will be removed from the code.

### PRs for API

* Must have unit tests
* Must have swagger documentation
* Should update `osd-sanity` with the new API.
