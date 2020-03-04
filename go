#!/usr/bin/env bash
export GOPRIVATE='*'

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
  echo "[INFO] download $pkgver" | tee /dev/stderr | cat >>/tmp/golog
  /bin/go mod download -json "$pkgver"
}

list(){
  pkg="$1"
  echo "[INFO] list $pkg" | tee /dev/stderr | cat >>/tmp/golog
  /bin/go list -m -versions -json "$pkg"
}

main(){
  if checkList "$@"; then
    pkg="$5"
    list "$pkg"
  elif checkDownload "$@"; then
    pkgver="$4"
    download "$pkgver"
  fi
}

main "$@"
