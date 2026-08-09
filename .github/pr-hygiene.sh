#!/usr/bin/env bash
#
# The deterministic pull request hygiene rules, and the fixtures that prove they
# bite.
#
# These are the rules that reason about the change rather than about the code.
# No analyser looks at whether a laboratory value moved without its provenance
# moving, because that is not a property of any file on its own: it is a property
# of the difference between two commits.
#
# The file is split in three on purpose. `facts` asks the hosting service what
# the pull request contains and writes it down as lines. `judge` reads those
# lines and decides, and it reads nothing else, so a rule can be run against a
# written-down pull request that never existed. `selftest` runs `judge` over the
# cases in .github/pr-hygiene/cases and refuses a verdict that is not the one
# recorded there.
#
# The split is what makes the rules testable at all. A rule living inside a
# workflow's `run:` block is exercised by nothing except the next real pull
# request, and this repository has already decided, in docs/gate-parity.md, that
# a check owes a demonstration that it refuses what it names.
#
# What that demonstration does not cover is `facts` itself. Nothing here runs
# `gh api`, so the step that turns a pull request into lines is judged by no
# fixture and only by the runs it makes. That is stated rather than worked
# around: the fixtures prove the rules, not the reading.

set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
usage:
  pr-hygiene.sh facts <output-file>       write the facts of this pull request
  pr-hygiene.sh judge <facts> <body>      apply the rules and report
  pr-hygiene.sh selftest                  run the rules against the recorded cases

facts needs GH_TOKEN, GH_REPO and PR_NUMBER in the environment.
judge writes one line per rule to stdout: "pass", "warn" or "fail", then the
rule's name, then what it saw. It exits 1 if any rule failed and 0 otherwise, so
a warning never reds a pull request. Set PR_HYGIENE_ANNOTATE=1 to also emit
GitHub annotations; the fixtures run without it so that what they compare is the
verdict and not the presentation.
USAGE
  exit 2
}

# ---------------------------------------------------------------------------
# facts
# ---------------------------------------------------------------------------

# A record is a JSON file under data/ that is not part of the schema. The schema
# is data and is not a record: it carries no laboratory value and has no
# provenance block, so the provenance rule has nothing to say about it. This is
# the near miss the rule is most likely to fire on wrongly, and it is a case in
# the fixtures rather than an assurance here.
is_record() {
  case "$1" in
    data/schema/*) return 1 ;;
    data/*.json) return 0 ;;
    data/*/*.json) return 0 ;;
    *) return 1 ;;
  esac
}

facts() {
  local out="${1:-}"
  [ -n "$out" ] || usage
  : >"$out"

  local base head
  base="$(gh api "repos/${GH_REPO}/pulls/${PR_NUMBER}" --jq '.base.sha')"
  head="$(gh api "repos/${GH_REPO}/pulls/${PR_NUMBER}" --jq '.head.sha')"
  if [ -z "$base" ] || [ -z "$head" ]; then
    echo "could not read the pull request's base and head" >&2
    return 1
  fi

  local changed
  changed="$(gh api --paginate "repos/${GH_REPO}/pulls/${PR_NUMBER}/files" --jq '.[].filename')"
  # An empty file list is a pull request that changes nothing, which is a thing
  # that can happen, so it is not an error. A failed call is, and the guard above
  # the assignment is what tells the two apart.

  local path
  while IFS= read -r path; do
    [ -n "$path" ] || continue
    printf 'changed\t%s\n' "$path" >>"$out"
    if is_record "$path"; then
      local before after
      # A file that did not exist at the base has no previous provenance to
      # compare against, so nothing is written for it and the rule below treats
      # it as an addition rather than as a change.
      if before="$(gh api "repos/${GH_REPO}/contents/${path}?ref=${base}" --jq '.content' 2>/dev/null | base64 -d | jq -cS '.provenance // null' 2>/dev/null)"; then
        after="$(gh api "repos/${GH_REPO}/contents/${path}?ref=${head}" --jq '.content' | base64 -d | jq -cS '.provenance // null')"
        if [ "$before" = "$after" ]; then
          printf 'provenance\t%s\tunchanged\n' "$path" >>"$out"
        else
          printf 'provenance\t%s\tchanged\n' "$path" >>"$out"
        fi
      fi
    fi
  done <<<"$changed"
}

# ---------------------------------------------------------------------------
# judge
# ---------------------------------------------------------------------------

# The one line a pull request body carries to say that a record changed without
# its provenance changing, and why. It is an argument attached to the change
# rather than a switch: the rule still fires, the body still has to say what it
# was, and a reviewer reads the reason beside the diff.
declare -r OVERRIDE_PREFIX='Record changed without its provenance: '

verdict_fail=0

report() {
  local level="$1" rule="$2" detail="$3"
  printf '%s\t%s\t%s\n' "$level" "$rule" "$detail"
  if [ "${PR_HYGIENE_ANNOTATE:-0}" = "1" ]; then
    case "$level" in
      fail) printf '::error::%s: %s\n' "$rule" "$detail" ;;
      warn) printf '::warning::%s: %s\n' "$rule" "$detail" ;;
    esac
  fi
  [ "$level" = "fail" ] && verdict_fail=1
  return 0
}

judge() {
  local factsfile="${1:-}" bodyfile="${2:-}"
  [ -n "$factsfile" ] && [ -n "$bodyfile" ] || usage
  [ -r "$factsfile" ] || { echo "cannot read $factsfile" >&2; return 1; }
  [ -r "$bodyfile" ] || { echo "cannot read $bodyfile" >&2; return 1; }

  local changed=() provenance_unchanged=()
  local kind path state
  while IFS=$'\t' read -r kind path state; do
    case "$kind" in
      changed) changed+=("$path") ;;
      provenance) [ "$state" = "unchanged" ] && provenance_unchanged+=("$path") ;;
      '') ;;
      *) echo "unknown fact: $kind" >&2; return 1 ;;
    esac
  done <"$factsfile"

  # Failing rule. A laboratory value whose source silently changes is the worst
  # thing that can happen to this database, and there is no ordinary change that
  # does it by accident: correcting a value, entering a remeasurement and
  # migrating a record all touch who entered it and when. The one shape that
  # legitimately does not is a schema migration that rewrites every record, and
  # that is what the body line is for.
  local offenders=() p
  for p in "${provenance_unchanged[@]:-}"; do
    [ -n "$p" ] || continue
    is_record "$p" && offenders+=("$p")
  done
  if [ "${#offenders[@]}" -gt 0 ]; then
    if grep -qF -- "$OVERRIDE_PREFIX" "$bodyfile"; then
      report pass record-provenance-fresh \
        "${#offenders[@]} record(s) changed with their provenance untouched, declared in the body"
    else
      report fail record-provenance-fresh \
        "$(printf '%s ' "${offenders[@]}")changed with the provenance block byte-identical; if that is intended, the body needs a line beginning '${OVERRIDE_PREFIX}'"
    fi
  else
    report pass record-provenance-fresh "no record changed with its provenance untouched"
  fi

  # Failing rule. CONTRIBUTING.md's first rule is that every change starts as an
  # issue, and a body naming no issue is not a borderline case: there is nothing
  # for a reviewer to read the change against.
  if grep -qE '#[0-9]+' "$bodyfile"; then
    report pass names-an-issue "the body references an issue"
  else
    report fail names-an-issue "the body references no issue"
  fi

  # Warning rule. A change to the arithmetic with no test change is usually a
  # gap and sometimes a comment, a rename or a documentation fix, so it
  # annotates. A check that reds a reasonable change gets ignored and then
  # switched off, and then it protects nothing at all.
  local calculation=() tests=()
  for p in "${changed[@]:-}"; do
    case "$p" in
      *_test.go) tests+=("$p") ;;
      acoustic/*.go|kernel/*.go|store/*.go) calculation+=("$p") ;;
    esac
  done
  if [ "${#calculation[@]}" -gt 0 ] && [ "${#tests[@]}" -eq 0 ]; then
    report warn calculation-changed-with-tests \
      "$(printf '%s ' "${calculation[@]}")changed and no test file did"
  else
    report pass calculation-changed-with-tests \
      "${#calculation[@]} calculation file(s) and ${#tests[@]} test file(s) changed"
  fi

  return "$verdict_fail"
}

# ---------------------------------------------------------------------------
# selftest
# ---------------------------------------------------------------------------

selftest() {
  local here cases failed=0 total=0
  here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  cases="$here/pr-hygiene/cases"
  [ -d "$cases" ] || { echo "no cases directory at $cases" >&2; return 1; }

  local factsfile name bodyfile expected got status
  for factsfile in "$cases"/*.facts; do
    [ -e "$factsfile" ] || { echo "no cases in $cases" >&2; return 1; }
    total=$((total + 1))
    name="$(basename "$factsfile" .facts)"
    bodyfile="$cases/$name.body"
    expected="$cases/$name.expect"
    if [ ! -r "$bodyfile" ] || [ ! -r "$expected" ]; then
      echo "FAIL  $name is missing its body or its expected verdict"
      failed=$((failed + 1))
      continue
    fi
    status=0
    got="$( (verdict_fail=0; judge "$factsfile" "$bodyfile") )" || status=$?
    if [ "$(printf 'exit=%s\n%s\n' "$status" "$got")" = "$(cat "$expected")" ]; then
      echo "ok    $name"
    else
      echo "FAIL  $name"
      echo "--- expected"
      cat "$expected"
      echo "--- got"
      printf 'exit=%s\n%s\n' "$status" "$got"
      failed=$((failed + 1))
    fi
  done

  if [ "$failed" -ne 0 ]; then
    echo "$failed of $total case(s) did not get the verdict recorded for them."
    return 1
  fi
  echo "$total case(s), every one judged as recorded."
}

case "${1:-}" in
  facts) shift; facts "$@" ;;
  judge) shift; judge "$@" ;;
  selftest) shift; selftest "$@" ;;
  *) usage ;;
esac
