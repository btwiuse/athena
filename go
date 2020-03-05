#!/usr/bin/env bash
export GOPRIVATE='*'
export GO=${GOBIN:-/bin/go}

checkDownload(){
  [[ "$1" == mod      ]] && \
  [[ "$2" == download ]] && \
  [[ "$3" == -json    ]] && \
  [[ "$4" =~ ^git.milvai.cn.*|^powerlaw.ai.* ]] && \
  true || false
}

checkList(){
  [[ "$1" == list      ]] && \
  [[ "$2" == -m        ]] && \
  [[ "$3" == -versions ]] && \
  [[ "$4" == -json     ]] && \
  [[ "$5" =~ ^git.milvai.cn.*|^powerlaw.ai.* ]] && \
  true || false
}

sync(){
  pkgver="$1"
  pkg="$(echo $pkgver | cut -d @ -f 1)"
  ver="$(echo $pkgver | cut -d @ -f 2)"
  cd "$(go env GOPATH)"/pkg/mod/cache/download/
  dst="$(getMod $pkgver | xargs head -n1 | sed -e 's,module ,,g')"

  dirFrom="$(getDir ${pkg}@${ver})"
  dirTo="$(getDir ${dst}@${ver})"
  mkdir -p $dirTo

  cp -r "$dirFrom"/* "$dirTo"

# zipFrom="$(getZip ${pkg}@${ver})"
  zipFile="$(basename $(getZip ${dst}@${ver}))"
  cd "$dirTo"
  unzip $zipFile
  mkdir -p $(dirname $dst)
  ln -sv ../../"${pkgver}" "${dst}@${ver}"
  zip -r tmp.zip "${dst}@${ver}"/*
  mv -f tmp.zip "$zipFile"
  rm -r powerlaw.ai git.milvai.cn
}

getInfo(){
  pkgver="$1"
  echo "$pkgver" | sed -e 's,@,/@v/,g' -e 's,$,.info,g'
}

getMod(){
  pkgver="$1"
  echo "$pkgver" | sed -e 's,@,/@v/,g' -e 's,$,.mod,g'
}

getZip(){
  pkgver="$1"
  echo "$pkgver" | sed -e 's,@,/@v/,g' -e 's,$,.zip,g'
}

getDir(){
  pkgver="$1"
  echo "$pkgver" | sed -e 's,@,/@v/,g' | xargs dirname
}

getList(){
  pkgver="$1"
  getDir "$pkgver" | sed -e 's,$,/list,g' 
}

download(){
  pkgver="$1"
  "$GO" mod download -json "$pkgver"
  if [[ "$pkgver" =~ ^git.milvai.cn.* ]]; then
    info "SYNC $pkgver"
    sync "$pkgver"
  fi
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
  echo "[GIAO] $@" | grep --color . 1>&2
}

list(){
  pkg="$1"
  if [[ "$pkg" =~ ^powerlaw.ai.* ]]; then
    info "=> $pkg"
    echo "$pkg" | sed s,powerlaw.ai,git.milvai.cn,g | xargs go list -m -versions -json | sed s,git.milvai.cn,powerlaw.ai,g
  else
    "$GO" list -m -versions -json "$pkg"
  fi
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
    info "LS $pkg"
    list "$pkg"
  elif checkDownload "$@"; then
    pkgver="$4"
    info "DL $pkgver"
    download "$pkgver"
  else
    info "GO $@"
    "$GO" "$@"
  fi
}

main "$@"
