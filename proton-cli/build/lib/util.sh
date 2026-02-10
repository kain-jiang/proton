# Wait for background jobs to finish. Return with
# an error status if any of the jobs failed.
function proton::util::wait-for-jobs {
  local fail=0
  local job
  for job in $(jobs -p); do
    wait "${job}" || fail=$((fail + 1))
  done
  return ${fail}
}
