#!/usr/bin/env bash

# highest number of servers and clients
NBR=${NBR:-3}
# Per default, have $NBR servers
NBR_SERVERS=${NBR_SERVERS:-$NBR}
# Per default, keep one inactive server
NBR_SERVERS_GROUP=${NBR_SERVERS_GROUP:-$(( NBR_SERVERS - 1))}
# Show the output of the commands: 0=none, 1=test-names, 2=all
DBG_TEST=${DBG_TEST:-0}
# DBG-level for server
DBG_SRV=${DBG_SRV:-0}
# APPDIR is usually where the test.sh-script is located
APPDIR=${APPDIR:-$(pwd)}
# The app is the name of the builddir
APP=${APP:-$(basename $APPDIR)}
# Name of conode-log
COLOG=conode

RUNOUT=$( mktemp )

# cleans the test-directory and builds the CLI binary
# Globals:
#   CLEANBUILD
#   APPDIR
#   APP
startTest(){
  set +m
  if [ "$CLEANBUILD" ]; then
    rm -f conode $APP
  fi
  build $APPDIR
}

# Prints `Name`, cleans up build-directory, deletes all databases from services
# in previous run, and calls `testName`.
# Arguments:
#   `Name` - name of the testing function to run
run(){
  cleanup
  echo -e "\n* Testing $1"
  sleep .5
  $1
}

# Asserts that the exit-code of running `$@` using `dbgRun` is `0`.
# Arguments:
#   $@ - used to run the command
testOK(){
  testOut "Assert OK for '$@'"
  if ! dbgRun "$@"; then
    fail "starting $@ failed"
  fi
}

# Asserts that the exit-code of running `$@` using `dbgRun` is NOT `0`.
# Arguments:
#   $@ - used to run the command
testFail(){
  testOut "Assert FAIL for '$@'"
  if dbgRun "$@"; then
    fail "starting $@ should've failed, but succeeded"
  fi
}

# Asserts `File` exists and is a file.
# Arguments:
#   `File` - path to the file to test
testFile(){
  testOut "Assert file $1 exists"
  if [ ! -f $1 ]; then
    fail "file $1 is not here"
  fi
}

# Asserts `File` DOES NOT exist.
# Arguments:
#   `File` - path to the file to test
testNFile(){
  testOut "Assert file $1 DOESN'T exist"
  if [ -f $1 ]; then
    fail "file $1 IS here"
  fi
}

# Asserts that `String` exists in `File`.
# Arguments:
#   `String` - what to search for
#   `File` - in which file to search
testFileGrep(){
  local G="$1" F="$2"
  testFile "$F"
  testOut "Assert file $F contains --$G--"
  if ! pcregrep -M -q "$G" $F; then
    fail "Didn't find '$G' in file '$F': $(cat $F)"
  fi
}

# Asserts that `String` is in the output of the command being run by `dbgRun`
# and all but the first input argument. Ignores the exit-code of the command.
# Arguments:
#   `String` - what to search for
#   `$@[1..]` - command to run
testGrep(){
  S="$1"
  shift
  testOut "Assert grepping '$S' in '$@'"
  runOutFile "$@"
  doGrep "$S"
  if [ ! "$EGREP" ]; then
    fail "Didn't find '$S' in output of '$@': $GREP"
  fi
}

# Asserts the output of the command being run by `dbgRun` and all but the first
# input argument is N lines long. Ignores the exit-code of the command.
# Arguments:
#   `N` - how many lines should be output
#   `$@[1..]` - command to run
testCountLines(){
  N="$1"
  shift
  testOut "Assert wc -l is $N lines in '$@'"
  runOutFile "$@"
  lines=`wc -l < $RUNOUT`
  if [ $lines != $N ]; then
    fail "Found $lines lines in output of '$@'"
  fi
}

# Asserts that `String` is NOT in the output of the command being run by `dbgRun`
# and all but the first input argument. Ignores the exit-code of the command.
# Arguments:
#   `String` - what to search for
#   `$@[1..]` - command to run
testNGrep(){
  G="$1"
  shift
  testOut "Assert NOT grepping '$G' in '$@'"
  runOutFile "$@"
  doGrep "$G"
  if [ "$EGREP" ]; then
    fail "DID find '$G' in output of '$@': $(cat $RUNOUT)"
  fi
}

# Asserts `String` is part of the last command being run by `testGrep` or
# `testNGrep`.
# Arguments:
#   `String` - what to search for
testReGrep(){
  G="$1"
  testOut "Assert grepping again '$G' in same output as before"
  doGrep "$G"
  if [ ! "$EGREP" ]; then
    fail "Didn't find '$G' in last output: $(cat $RUNOUT)"
  fi
}

# Asserts `String` is NOT part of the last command being run by `testGrep` or
# `testNGrep`.
# Arguments:
#   `String` - what to search for
testReNGrep(){
  G="$1"
  testOut "Assert grepping again NOT '$G' in same output as before"
  doGrep "$G"
  if [ "$EGREP" ]; then
    fail "DID find '$G' in last output: $(cat $RUNOUT)"
  fi
}

# used in test*Grep methods.
doGrep(){
  # echo "grepping in $RUNOUT"
  # cat $RUNOUT
  WC=$( cat $RUNOUT | egrep "$1" | wc -l )
  EGREP=$( cat $RUNOUT | egrep "$1" )
}

# Asserts that `String` exists exactly `Count` times in the output of the
# command being run by `dbgRun` and all but the first two arguments.
# Arguments:
#   `Count` - number of occurences
#   `String` - what to search for
#   `$@[2..]` - command to run
testCount(){
  C="$1"
  G="$2"
  shift 2
  testOut "Assert counting '$C' of '$G' in '$@'"
  runOutFile "$@"
  doGrep "$G"
  if [ $WC -ne $C ]; then
    fail "Didn't find '$C' (but '$WC') of '$G' in output of '$@': $(cat $RUNOUT)"
  fi
}


# Outputs all arguments if `DBT_TEST -ge 1`
# Globals:
#   DBG_TEST - determines debug-level
testOut(){
  if [ "$DBG_TEST" -ge 1 ]; then
    echo -e "$@"
  fi
}

# Outputs all arguments if `DBT_TEST -ge 2`
# Globals:
#   DBG_TEST - determines debug-level
dbgOut(){
  if [ "$DBG_TEST" -ge 2 ]; then
    echo -e "$@"
  fi
}

# Runs `$@` and outputs the result of `$@` if `DBG_TEST -ge 2`. Redirects the
# output in all cases if `OUTFILE` is set.
# Globals:
#   DBG_TEST - determines debug-level
#   OUTFILE - if set, used to write output
dbgRun(){
  if [ "$DBG_TEST" -ge 2 ]; then
    OUT=/dev/stdout
  else
    OUT=/dev/null
  fi
  if [ "$OUTFILE" ]; then
    "$@" 2>&1 | tee $OUTFILE > $OUT
  else
    "$@" 2>&1 > $OUT
  fi
}

runGrepSed(){
  GREP="$1"
  SED="$2"
  shift 2
  runOutFile "$@"
  doGrep "$GREP"
  SED=$( echo $EGREP | sed -e "$SED" )
}

runOutFile(){
  OLDOUTFILE=$OUTFILE
  OUTFILE=$RUNOUT
  dbgRun "$@"
  OUTFILE=$OLDOUTFILE
}

fail(){
  echo
  echo -e "\tFAILED: $@"
  cleanup
  exit 1
}

backg(){
  ( "$@" 2>&1 & )
}

# Builds the app stored in the directory given in the first argument.
# Globals:
#   CLEANBUILD - if set, forces build of app, even if it exists.
#   TAGS - what tags to use when calling go build
# Arguments:
#   builddir - where to search for the app to build
build(){
  local builddir=$1
  local app=$( basename $builddir )
  if [ ! -e $app -o "$CLEANBUILD" ]; then
    testOut "Building $app"
    if ! go build -o $app $TAGS $builddir/*.go; then
      fail "Couldn't build $builddir"
    fi
  else
    dbgOut "Not building $app because it's here"
  fi
}

buildDir(){
  BUILDDIR=${BUILDDIR:-$(mktemp -d)}
  mkdir -p $BUILDDIR
  testOut "Working in $BUILDDIR"
  cd $BUILDDIR
}

# Magical method that tries very hard to build a conode. If no arguments given,
# it will search for a service first in the `./service` directory, and if not
# found, in the `../service` directory.
# If a directory is given as an argument, the service will be taken from that
# directory.
# Globals:
#   APPDIR - where the app is stored
# Arguments:
#   [serviceDir, ...] - if given, used as directory to be included. More than one
#                       argument can be given.
buildConode(){
  local incl="$@"
  gopath=`go env GOPATH`
  if [ -z "$incl" ]; then
    echo "buildConode: No import paths provided. Searching."
    for i in service ../service
    do
      if [ -d $APPDIR/$i ]; then
        local pkg=$( realpath $APPDIR/$i | sed -e "s:$gopath/src/::" )
        incl="$incl $pkg"
      fi
    done
    echo "Found: $incl"
  fi

  local cotdir=$( mktemp -d )/conode
  mkdir -p $cotdir

  ( echo -e "package main\nimport ("
    for i in $incl; do
      echo -e "\t_ \"$i\""
    done
  echo ")" ) > $cotdir/import.go

  if [ ! -f "$gopath/src/github.com/dedis/cothority/conode/conode.go" ]; then
    echo "Cannot find package github.com/dedis/cothority."
    exit 1
  fi
  cp "$gopath/src/github.com/dedis/cothority/conode/conode.go" $cotdir/conode.go

  build $cotdir
  rm -rf $cotdir
  setupConode
}

setupConode(){
  # Don't show any setup messages
  DBG_OLD=$DBG_TEST
  DBG_TEST=0
  rm -f public.toml
  for n in $( seq $NBR_SERVERS ); do
    co=co$n
    rm -f $co/*
    mkdir -p $co
    echo -e "localhost:200$(( 2 * $n ))\nCot-$n\n$co\n" | dbgRun runCo $n setup
    if [ ! -f $co/public.toml ]; then
      echo "Setup failed: file $co/public.toml is missing."
      exit
    fi
    if [ $n -le $NBR_SERVERS_GROUP ]; then
      cat $co/public.toml >> public.toml
    fi
  done
  DBG_TEST=$DBG_OLD
}

runCoBG(){
  for nb in "$@"; do
    dbgOut "starting conode-server #$nb"
    (
      # Always redirect output of server in log-file, but
      # only output to stdout if DBG_SRV > 0.
      rm -f "$COLOG$nb.log.dead"
      if [ "$DBG_SRV" = 0 ]; then
        ./conode -d $DBG_SRV -c co$nb/private.toml server > "$COLOG$nb.log" | cat
      else
        ./conode -d $DBG_SRV -c co$nb/private.toml server 2>&1 | tee "$COLOG$nb.log"
      fi
      touch "$COLOG$nb.log.dead"
    ) &
  done
  sleep 1
  for nb in "$@"; do
    dbgOut "checking conode-server #$nb"
    if [ -f "$COLOG$nb.log.dead" ]; then
      echo "Server $nb failed to start:"
      cat "$COLOG$nb.log"
      exit 1
    fi
  done
}

runCo(){
  local nb=$1
  shift
  dbgOut "starting conode-server #$nb"
  dbgRun ./conode -d $DBG_SRV -c co$nb/private.toml "$@"
}

cleanup(){
  pkill -9 conode 2> /dev/null
  pkill -9 ^${APP}$ 2> /dev/null
  sleep .5
  rm -f co*/*bin
  rm -f cl*/*bin
  if [ -z "$KEEP_DB" ]; then
    rm -rf $CONODE_SERVICE_PATH
  fi
}

stopTest(){
  cleanup
  if [ $( basename $BUILDDIR ) != build ]; then
    dbgOut "removing $BUILDDIR"
    rm -rf $BUILDDIR
  fi
  echo "Success"
}

if ! which pcregrep > /dev/null; then
  echo "*** WARNING ***"
  echo "Most probably you're missing pcregrep which might be used here..."
  echo "On mac you can install it with"
  echo -e "\n  brew install pcre\n"
  echo "Not aborting because it might work anyway."
  echo
fi

if ! which realpath > /dev/null; then
  echo "*** WARNING ***"
  echo "Most probably you're missing realpath which might be used here..."
  echo "On mac you can install it with"
  echo -e "\n  brew install coreutils\n"
  echo "Not aborting because it might work anyway."
  echo
  realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
  }
fi

for i in "$@"; do
  case $i in
    -b|--build)
      CLEANBUILD=yes
      shift # past argument=value
      ;;
    -nt|--notemp)
      BUILDDIR=$(pwd)/build
      shift # past argument=value
      ;;
  esac
done
buildDir

export CONODE_SERVICE_PATH=$BUILDDIR/service_storage
