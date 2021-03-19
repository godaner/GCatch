if test -f "$Z3"; then
  echo "Z3 exists"
else
  echo "Installing Z3 to a default position, normally $Z3"
  echo "Could fail if ran without sudo"
  cd ./tools/z3
  python scripts/mk_make.py
  cd build
  make
  sudo make install
  sudo cp ./libz3.so /lib/
  sudo cp ./libz3.so ./build/
  sudo cp ./z3 /usr/bin/
  sudo cp ./z3 /usr/local/bin/
  sudo cp ./z3 ./build/
fi
