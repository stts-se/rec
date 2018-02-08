from __future__ import division
import sys, os, wave


kaldi_dir = "~/git/kaldi/egs/irish_test_dec08"
model_dir = "exp/tri2b_mpe/"
recserver_dir = "~/git/rec/recserver"

def decode(wavfile):
    #check audio
    wav = wave.open(wavfile)
    rate = wav.getframerate()
    ch = wav.getnchannels()
    if rate != 16000 or ch != 1:
        return "ERROR: wavfile %s is %d channel %d Hz" % (wavfile, ch, rate)
    frames = wav.getnframes()
    dur = frames/rate
    #print("Duration: %.2f s" % dur)
    
    #save audio in tmp?

    #call decode script
    decode_cmd = "cd %s; bash decode_hb_2.sh %s/%s %s" % (kaldi_dir, recserver_dir, wavfile, model_dir)
    sys.stderr.write(decode_cmd+"\n")
    result = os.popen(decode_cmd).read()

    #return output
    return result.strip()






if __name__ == "__main__":
    wavfile = sys.argv[1]
    
    #wavfile = "/home/harald/test_audio/1_2_3_co_hts_16.wav"

    print(decode(wavfile))
