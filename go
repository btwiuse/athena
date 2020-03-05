#!/usr/bin/env bash
export GOPRIVATE='*'
export GO=${GOBIN:-/bin/go}

checkDownload(){
  [[ "$1" == mod      ]] && \
  [[ "$2" == download ]] && \
  [[ "$3" == -json    ]] && \
  true || false
}

checkList(){
  [[ "$1" == list      ]] && \
  [[ "$2" == -m        ]] && \
  [[ "$3" == -versions ]] && \
  [[ "$4" == -json     ]] && \
  true || false
}

download(){
  pkgver="$1"
  "$GO" mod download -json "$pkgver"
}

# mod O -> K
# mod O release -> go mod download
# mod O list -> go mod replicate
# mod K list -> copy mod O to mod K
#
# case mod path == module name:
#   lookup parent mod (parent file)
#     ok:
#       virtual mod
#     not ok:
#       ordinary mod
#
# case mod path != module name:
#   checkout virtual mod

info(){
  echo "[INFO] $@" | grep --color .
}

list(){
  pkg="$1"
  "$GO"/bin/go list -m -json "$pkg"
# if /bin/go list -m -json all | jq -r .Path | grep -q "$pkg"; then
#   # find parent module
#   parent="cat $(/bin/go list -m -json "$pkg" | jq -r .Path)/parent"
#   /bin/go list -m -json "$parent"
# fi
# # mod O -> checkout K also
# /bin/go list -m -versions -json git.milvai.cn/platform/gotools | jq -r .Path
# /bin/go list -m -versions -json git.milvai.cn/platform/gotools | jq -r .Dir | xargs -I@ grep -e ^module @/go.mod | sed -e 's,module ,,g'
}

main(){
  if checkList "$@"; then
    pkg="$5"
    info "LIST $pkg"
    list "$pkg"
  elif checkDownload "$@"; then
    pkgver="$4"
    info "DOWN $pkgver"
    download "$pkgver"
  else
    info "GO $@"
    "$GO" "$@"
  fi
}

main "$@"
