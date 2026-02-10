readonly COVERAGE_REPORT_COBERTURA_XML_PATH="${COVERAGE_REPORT_COBERTURA_XML_PATH-cover.xml}"

# Ensure the gocov and gocov-xml exists.
function proton::gocov::verify {
  if [[ -z "$(command -v gocov)" ]]; then
    proton::log::usage_from_stdin <<EOF
  Can't find 'gocov' in PATH, please fix and retry.
  See https://github.com/axw/gocov#installation for installation instructions.
EOF
      return 2
  fi

  if [[ -z "$(command -v gocov-xml)" ]]; then
    proton::log::usage_from_stdin <<EOF
  Can't find 'gocov-xml' in PATH, please fix and retry.
  See https://github.com/AlekSi/gocov-xml#installation for installation instructions.
EOF
      return 2
  fi
}

# Run unit testing and write a junit file.
function proton::gocov::convert {
  gocov convert cover.out | gocov-xml > "${COVERAGE_REPORT_COBERTURA_XML_PATH}"
}
