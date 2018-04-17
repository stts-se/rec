cmd=$0
dir=`dirname $cmd`

if [ $# -lt 1 ]; then
    echo "Usage: $cmd <wavfiles>" 1>&2
    exit 1
fi

cd $dir

for wav in $*; do
   play $wav &> /dev/null
   go run reccli.go -u tmptestxx $wav
done
