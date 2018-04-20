git=`ls -d $HOME/git* | egrep "^.*/(git.*repos|git)$" | head -1`

cd $git/e-lexia/pocketsphinx

gunicorn -b localhost:8000 demo_server:api nst_10000_adapted_467 &
gunicorn -b localhost:9090 demo_server:api nst_46912 &
gunicorn -b localhost:9091 demo_server:api nst_46912_adapted_467 &
gunicorn -b localhost:9999 demo_server:api elexia_448 &
