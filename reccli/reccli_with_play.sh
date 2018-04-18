cmd=$0
wd=`pwd`
dir=`dirname $cmd`

if [ $# -lt 1 ]; then
    echo "Usage: $cmd <wavfiles>" 1>&2
    exit 1
fi

for wav in $*; do
    cd $wd
    abswav=`realpath $wav`
    play $wav &> /dev/null
    cd $dir && go run reccli.go -u tmptestxx $abswav
done
