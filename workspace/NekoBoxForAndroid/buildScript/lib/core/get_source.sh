#!/bin/bash
set -e

source "buildScript/init/env.sh"
ENV_NB4A=1
source "buildScript/lib/core/get_source_env.sh"
BYEDPI_PATCH=$(realpath "patches/byedpi/0001-byedpi-jni-api.patch")
pushd ..

# A prepared workspace is committed as plain source directories without
# nested .git metadata. In that case the sources are already pinned and
# patched, so do not run git checkout/apply against the outer application
# repository. For ordinary local builds with cloned dependency repos, keep
# the original pinning behavior.
checkout_if_repo() {
  local dep="$1" commit="$2"
  if git -C "$dep" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    if [ "$(git -C "$dep" rev-parse HEAD)" != "$commit" ]; then
      git -C "$dep" checkout "$commit"
    fi
  else
    echo "$dep is a prepared source directory; skipping git checkout"
  fi
}

#### sing-box
if [ ! -d "sing-box" ]; then
  git clone --no-checkout {sing-box-repo-here} sing-box
fi
checkout_if_repo sing-box "$COMMIT_SING_BOX"

#### sing-vmess
if [ ! -d "sing-vmess" ]; then
  git clone --no-checkout {sing-vmess-repo-here} sing-vmess
fi
checkout_if_repo sing-vmess "$COMMIT_SING_VMESS"

#### libneko
if [ ! -d "libneko" ]; then
  git clone --no-checkout {libneko-repo-here} libneko
fi
checkout_if_repo libneko "$COMMIT_LIBNEKO"

#### byedpi
if [ ! -d "byedpi" ]; then
  git clone --no-checkout https://github.com/hufrea/byedpi.git byedpi
fi
if git -C byedpi rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  if [ "$(git -C byedpi rev-parse HEAD)" != "$COMMIT_BYEDPI" ]; then
    git -C byedpi checkout "$COMMIT_BYEDPI"
    git -C byedpi apply "$BYEDPI_PATCH"
  fi
else
  echo "byedpi is a prepared source directory; skipping git checkout/apply"
fi

####

popd
