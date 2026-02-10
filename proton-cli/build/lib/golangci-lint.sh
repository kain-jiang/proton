readonly REPORT_JUNIT_XML_PATH="${REPORT_JUNIT_XML_PATH-golangci-lint-report.xml}"

# Ensure the golangci-lint exists.
function proton::golangci-lint::verify_version {
  if [[ -z "$(command -v golangci-lint)" ]]; then
    proton::log::usage_from_stdin <<EOF
  Can't find 'golangci-lint' in PATH, please fix and retry.
  See https://golangci-lint.run/usage/install for installation instructions.
EOF
      return 2
  fi
}

# Run lint and output as junit-xml
function proton::golangci-lint::run::junit-xml {
  golangci-lint run --out-format=junit-xml >"${REPORT_JUNIT_XML_PATH}"
}
