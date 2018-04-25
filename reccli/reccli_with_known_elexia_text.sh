cmd=`basename $0`
wd=`pwd`
dir=`dirname $0`

verbose=0

url="http://localhost:9993/rec/process/?verb=true"

show_help() {
    echo "Usage: $cmd -[vh] <wavfiles>" 1>&2
    echo "       -h print this help and exit" 1>&2
    echo "       -u URL for calling rec server (default: $url)" 1>&2
    echo "       -v verbose output (with complete json input and result)" 1>&2
}


while getopts "hu:v" opt; do
    case "$opt" in
    h)
        show_help
        exit 0
        ;;
    v)  verbose=1
        ;;
    u)  url=$OPTARG
        ;;
    esac
done

shift $((OPTIND-1))

if [ $# -lt 1 ]; then
    show_help
    exit 0
fi

outprefix=""
if [ $verbose -eq 1 ]; then
    outprefix="[$cmd]\t"
fi

run_reccli() {
    uid=$(uuidgen)
    cd $dir
    target_text=$1
    tmpfile="/tmp/reccli-$uid.out"
    reccli_cmd="go run reccli.go -u elexia_test -url $url -t \"$target_text\" $abswav" 1>&2
    if go run reccli.go -u elexia_test -url $url -t "$target_text" $abswav > $tmpfile; then
	result=`cat $tmpfile | egrep recognition_result | head -1 | sed 's/.*recognition_result": *"\([^"]*\)",.*$/\1/'`
	echo "$result"
	if [ $verbose -eq 1 ]; then
	    echo "-- verbose output --"
	    echo $tmpfile
	fi
    else
	echo $reccli_cmd
	cat $tmpfile
	rm $tmpfile
	return 1
    fi
}

for wav in $*; do
    cd $wd
    basename=`basename $wav | sed 's/.wav//'`
    abswav=`realpath $wav`
    wavdir=`dirname $abswav`
    jsondir=`realpath $wavdir/../json`
    jsonfile=$jsondir/${basename}.json
    if [ ! -e $jsonfile ]; then
	echo "JSON file doesn't exist: $jsonfile" 1>&2
	exit 1
    fi
    target_text=`cat $jsonfile | egrep 'target_utt' | sed 's/.*": "\([^"]*\)".*/\1/'`
    if resp=$(run_reccli $target_text); then
	result=`echo $resp | sed 's/\s-- verbose output.*//'`
	tmpfile=`echo $resp | sed 's/.*verbose output --//'`
	correct="NO"
	if [[ $result == $target_text ]]; then
	    correct="YES"
	fi
	printf "$outprefix$wav\t$target_text\t$result\t$correct\n"
	if [ $verbose -eq 1 ]; then
	    echo "-- verbose output --"
	    cat $tmpfile
	fi
    else
	echo "FAILED $wav" 1>&2
	exit 1
    fi
    sleep 1
done
