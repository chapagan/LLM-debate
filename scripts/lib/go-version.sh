required_go_version() {
  printf '%s\n' "1.26.4"
}

go_version_at_least() {
  current=$1
  required=$2

  current_major=${current%%.*}
  current_rest=${current#*.}
  current_minor=${current_rest%%.*}
  current_patch=${current_rest#*.}
  [ "$current_patch" != "$current_rest" ] || current_patch=0

  required_major=${required%%.*}
  required_rest=${required#*.}
  required_minor=${required_rest%%.*}
  required_patch=${required_rest#*.}
  [ "$required_patch" != "$required_rest" ] || required_patch=0

  [ "$current_major" -gt "$required_major" ] && return 0
  [ "$current_major" -lt "$required_major" ] && return 1
  [ "$current_minor" -gt "$required_minor" ] && return 0
  [ "$current_minor" -lt "$required_minor" ] && return 1
  [ "$current_patch" -ge "$required_patch" ]
}

require_go_version() {
  required=$(required_go_version)

  command -v go >/dev/null 2>&1 || {
    echo "Go is required. Install Go $required or newer, then rerun this script." >&2
    exit 1
  }

  current=$(go env GOVERSION 2>/dev/null | sed 's/^go//')
  if [ -z "$current" ]; then
    current=$(go version | awk '{print $3}' | sed 's/^go//')
  fi

  if ! go_version_at_least "$current" "$required"; then
    echo "Go $required or newer is required. Found Go $current." >&2
    echo "Download the latest Go 1.26 release from https://go.dev/dl/." >&2
    exit 1
  fi
}
