cmd=`basename $0`
wd=`pwd`
dir=`dirname $0`

verbose=0

show_help() {
    echo "Usage: $cmd -[vh] <wavfiles>" 1>&2
    echo "       -h print this help and exit" 1>&2
    echo "       -v verbose output (with complete json input and result)" 1>&2
}

while getopts "hv" opt; do
    case "$opt" in
    h)
        show_help
        exit 0
        ;;
    v)  verbose=1
        ;;
    esac
done

shift $((OPTIND-1))

if [ $# -lt 1 ]; then
    show_help
    exit 0
fi

run_reccli() {
    uid=$(uuidgen)
    cd $dir
    target_text=$1
    tmpfile="/tmp/reccli-$uid.out"
    reccli_cmd="go run reccli.go -u elexia_test -t \"$target_text\" $abswav" 1>&2
    if go run reccli.go -u elexia_test -t "$target_text" $abswav >& $tmpfile; then
	result=`cat $tmpfile | egrep recognition_result | head -1 | sed 's/.*recognition_result": *"\([^"]*\)",.*$/\1/'`
	echo "$result"
	if [ $verbose -eq 1 ]; then
	    cat $tmpfile
	fi
	rm $tmpfile
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
    echo "WAV:     $abswav"
    echo "JSON:    $jsonfile"
    echo "TARGET:  <$target_text>"
    if result=$(run_reccli $target_text); then
	correct="NO"
	if [[ $result == $target_text ]]; then
	    correct="YES"
	fi
	echo "RESULT:  <$result>"
	echo "CORRECT: $correct"
	echo ""
    else
	echo "FAILED:  $result"
	exit 1
    fi
done
