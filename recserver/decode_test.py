from __future__ import division
import sys, os, wave


kaldi_dir = "~/git/kaldi/egs/irish_test_dec08"
model_dir = "exp/tri2b_mpe/"
recserver_dir = "~/git/rec/recserver"

def decode(wavfile):
    
    # TODO REMOVE
    return "DUMMY RETURN VALUE"
    
    if not os.path.isdir(kaldi_dir):
        sys.stderr.write("kaldi_dir doesn't exist : "+ kaldi_dir +"\n")
        exit(1)
        return

    m_dir = kaldi_dir +"/"+ model_dir
    if not os.path.isdir(m_dir):
        sys.stderr.write("model_dir doesn't exist : "+ m_dir +"\n")
        exit(1)
        return
    
    if not os.path.isdir(recserver_dir):
        sys.stderr.write("recserver_dir doesn't exist : "+ recserver_dir + "\n")
        exit(1)
        return
    
        
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
    result0 = os.popen(decode_cmd)
    result = result0.read()

    rc = result0.close()
    if rc is not None and rc >> 8:
        sys.stderr.write("Command failed : " + decode_cmd + "\n")
        exit(1)
        
    #return output
    return result.strip()






if __name__ == "__main__":
    wavfile = sys.argv[1]
    
    #wavfile = "/home/harald/test_audio/1_2_3_co_hts_16.wav"

    print(decode(wavfile))
