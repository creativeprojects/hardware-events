#!/bin/sh

cleanInstall() {
    printf "\033[32m Post Install of a clean install\033[0m\n"
}

upgrade() {
    printf "\033[32m Post Install of an upgrade\033[0m\n"
}

action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
  # Alpine linux does not pass args, and deb passes $1=configure
  action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
    # deb passes $1=configure $2=<current version>
    action="upgrade"
fi

case "$action" in
  "1" | "install")
    printf "\033[32m Post Install of a clean install\033[0m\n"
    cleanInstall
    ;;
  "2" | "upgrade")
    printf "\033[32m Post Install of an upgrade\033[0m\n"
    upgrade
    ;;
  *)
    # $1 == version being installed  
    printf "\033[32m Alpine\033[0m"
    cleanInstall
    ;;
esac