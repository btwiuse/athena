#!/usr/bin/env bash
# export GOPRIVATE='*'
export GO="${GOBIN:-/bin/go}"
export SELF="$(realpath "$0")"

checkDownload(){
  [[ "$1" == "mod" ]] && \
  [[ "$2" == "download" ]] && \
  [[ "$3" == "-json" ]] && \
  [[ "$4" =~ ^git.milvai.cn.*|^powerlaw.ai.* ]] && \
  true || false
}

checkList(){
  [[ "$1" == "list" ]] && \
  [[ "$2" == "-m" ]] && \
  [[ "$3" == "-versions" ]] && \
  [[ "$4" == "-json" ]] && \
  [[ "$5" =~ ^git.milvai.cn.*|^powerlaw.ai.* ]] && \
  true || false
}

sync(){
  pkgver="$1"
  pkg="$(echo "$pkgver" | cut -d @ -f 1)"
  ver="$(echo "$pkgver" | cut -d @ -f 2)"
  cd "$($GO env GOPATH)/pkg/mod/cache/download/" || exit
  dst="$(getMod "$pkgver" | xargs head -n1 | sed -e 's,module ,,g')"

  dirFrom="$(getDir "${pkg}@${ver}")"
  dirTo="$(getDir "${dst}@${ver}")"
  mkdir -p "$dirTo"

  zipFilePath="$(getZip "${dst}@${ver}")"
  if [[ -s "$zipFilePath" ]]; then
    return
  fi

  cp -r "$dirFrom"/* "$dirTo"
  echo "${pkg}" > "${dirTo}"/parent

# zipFrom="$(getZip ${pkg}@${ver})"
  zipFile="$(basename "${zipFilePath}")"
  cd "$dirTo" || exit
  unzip "$zipFile"
  mkdir -p "$(dirname "$dst")"
  ln -sv "$(relpath "${dst}")${pkgver}" "${dst}@${ver}"
  zip -r -D -m tmp.zip "${dst}@${ver}"/*
  # find "${dst}@${ver}"/ -type f | tee -a /tmp/asdf | xargs -L1 zip tmp.zip
  mv -f tmp.zip "$zipFile"
  rm -r powerlaw.ai git.milvai.cn
}

relpath(){
  echo "$1" | grep -o / | sed s,/,../,g | tr -d '\n'
}

getPkgParent(){
  pkg="$1"
  echo "$pkg"@ | sed -e 's,@,/@v/,g' -e 's,$,parent,g'
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
  "$GO" mod download -x -json "$pkgver"
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
  echo "[GIAO]" "$@" | grep --color . 1>&2
  # echo "[GIAO]" "$@" | grep --color . | tee -a /tmp/golog | cat 1>&2
}

list(){
  pkg="$1"
  if [[ "$pkg" =~ ^powerlaw.ai.* ]]; then
    cd "$($GO env GOPATH)"/pkg/mod/cache/download/ || exit
    parentFile="$(getPkgParent "$pkg")"
    if [[ -f parentFile ]]; then
      : # TODO: report to sentry: please manually run `go list -m git.milvzi.cn/parent/package` to register new powerlaw.ai package
    fi
    parent="$(cat "$parentFile")"
    # parent="$(find git.milvai.cn -name '*.mod' | xargs -L1 -I@ bash -c 'echo @ $(head -n1 @ | sed -e "s,module ,,g")' | grep "$pkg"$ | cut -d @ -f 1 | head -n1 | sed s,/$,,g)"
    info "$pkg => $parent"
    cd $OLDPWD
    # here call go itself TODO: call function instead
    "$SELF" list -m -versions -json "$parent" | sed -e "s,$parent,$pkg,g" # TODO: ensure $pkg.info exist
  elif [[ "$pkg" =~ ^git.milvai.cn.* ]]; then
    versionFile="$(mktemp)"
    "$GO" list -x -m -versions -json "$pkg" | tee -a "$versionFile"
    # TODO: check if at least one .mod file exists for each git.milvai.cn package
    # here call go itself TODO: call function instead
    version="$(cat $versionFile | jq -r .Version)"
    info "version=$version GO=$(which go) SELF=$SELF"
    if ! [[ -s $($GO env GOPATH)/pkg/mod/cache/download/$(getZip "${pkg}@${version}") ]]; then
      "$SELF" mod download -json "$pkg"@"$version" | grep --color . 1>&2
    fi
  else
    "$GO" list -x -m -versions -json "$pkg"
  fi
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
    info "go" "$@"
    "$GO" "$@"
  fi
}

main "$@"
