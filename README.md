# Nexus Resource

![](https://github.com/trecnoc/nexus-resource/workflows/CI/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/trecnoc/nexus-resource)](https://goreportcard.com/report/github.com/trecnoc/nexus-resource)

Versions objects in a Nexus repository of type Raw only, by pattern-matching
filenames to identify version numbers.

## Source Configuration

* `url`: *Required.* The url of the Nexus server.

* `repository`: *Required.* The name of the repository.

* `username`: *Required.* The username for access the repository.

* `password`: *Required.* The password for access the repository.

* `group`: *Required for check and out.* The repository artifact group, supports
  Glob patterns for `check`.

* `regexp`: *Required.* The pattern to match artifact name against within Nexus;
  this regex should match the full name of the files, which consists of the
  `group` minus the leading '/'. The first grouped match is used to extract the
  version, or if a group is explicitly named `version`, that group is used. At
  least one capture group must be specified, with parentheses.

  The version extracted from this pattern is used to version the resource.
  Semantic versions, or just numbers, are supported. Accordingly, full regular
  expressions are supported, to specify the capture groups.

* `timeout`: *Optional defaults to `10`.* Timeout for the internal HTTP Client in
  seconds.

* `debug`: *Optional defaults to `false`.* Debug flag for enabling logging and
  request file output in `/tmp`.

## Behavior

### `check`: Extract versions from the repository.

Artifacts will be found via the pattern configured by `regexp` in the provided
`group`. The versions will be used to order them (using [semver](http://semver.org/)).
Each artifact's filename is the resulting version.

### `in`: Fetch an artifact from the repository.

Places the following files in the destination:

* `(filename)`: The file fetched from the repository.

* `sha`: A file containing the SHA of the artifact.

* `url`: A file containing the URL of the artifact.

* `version`: The version identified in the file name.

#### Parameters

* `skip_download`: *Optional.* Defaults to `false`. Skip downloading object from
  Nexus. Value need to be a true/false string.

* `unpack`: *Optional.* Defaults to `false`. If true and the file is an archive
  (tar, gzipped tar, other gzipped file, or zip), unpack the file. Gzipped
  tarballs will be both ungzipped and untarred.

### `out`: Upload an object to the repository.

Given a file specified by `file`, upload it to the Nexus repository in the
provided `group`.

#### Parameters

* `file`: *Required.* Path to the file to upload, provided by an output of a task.
  If multiple files are matched by the glob, an error is raised. The matching
  syntax is bash glob expansion, so no capture groups, etc.

## Example Configuration

### Resource

When the file has the version name in the filename

``` yaml
- name: release
  type: nexus
  source:
    url: http://127.0.0.1
    repository: repositoryName
    group: /path/to
    regexp: path/to/release-(.*).tgz
```

### Plan

``` yaml
- get: release
```

``` yaml
- put: release
  params:
    file: path/to/release-*.tgz
```

## Developing on this resource

First get the resource via:
`go get github.com/trecnoc/nexus-resource`

## Development

### Prerequisites

* golang is *required* - version 1.14.x is tested; earlier versions may also
  work.
* docker is *required* - version 19.03.x is tested; earlier versions may also
  work.

### Local

Generate the Fakes with Counterfeiter if running tests locally or use the provided
scrips in the `scripts` folder.

Counterfeiter can be run with `go generate ./...`

### Docker

#### Running the tests

The tests have been embedded with the `Dockerfile`; ensuring that the testing
environment is consistent across any `docker` enabled platform. When the docker
image builds, the test are run inside the docker container, on failure they
will stop the build.

Run the tests with the following commands:

```sh
docker build -t nexus-resource -f Dockerfile .
```

##### Integration tests

The integration requires access to a Nexus server with a Raw repository.
The `docker build` step requires setting `--build-args` so the integration will run.

Run the tests with the following command:

```sh
docker build . -t nexus-resource -f Dockerfile \
  --build-arg NEXUS_TESTING_URL="some-url" \
  --build-arg NEXUS_TESTING_USERNAME="some-username" \
  --build-arg NEXUS_TESTING_PASSWORD="some-password" \
  --build-arg NEXUS_TESTING_REPOSITORY="some-repository"
```

### Contributing

Please make all pull requests to the `master` branch and ensure tests pass
locally.
