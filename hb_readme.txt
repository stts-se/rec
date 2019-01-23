To start basic recognition server for Irish:

cd /home/harald/git/kaldi/egs/irish_test_dec08/
nohup bash start_decode_test_server.sh &

cd /home/harald/git/rec/recserver
nohup go run *.go ../config/config-irish.json


URL:
http://www.abair.tcd.ie/rec/
