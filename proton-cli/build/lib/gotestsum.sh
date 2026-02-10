readonly UNIT_TESTING_REPORT_JUNIT_XML_PATH="${UNIT_TESTING_REPORT_JUNIT_XML_PATH-unit-tests.xml}"

# Ensure the gotestsum exists.
function proton::gotestsum::verify {
  if [[ -z "$(command -v gotestsum)" ]]; then
    proton::log::usage_from_stdin <<EOF
  Can't find 'gotestsum' in PATH, please fix and retry.
  See https://github.com/gotestyourself/gotestsum#install for installation instructions.
EOF
      return 2
  fi
}

# Run unit testing and write a junit file.
function proton::gotestsum::junitfile {
  gotestsum --junitfile="${UNIT_TESTING_REPORT_JUNIT_XML_PATH}" -- -coverprofile cover.out -gcflags -l ./...
}
