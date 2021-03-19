CURDIR=`pwd`
Z3=/usr/local/bin/z3
if test -f "$Z3"; then
  echo "Z3 exists"
else
  echo "Z3 is not installed in $Z3. Please run installZ3.sh with sudo or checkout https://github.com/Z3Prover/z3 to install Z3"
  exit 1
fi
cd $CURDIR
CGO_ENABLED=1 go install
cp $GOBIN/GCatch ./build/
