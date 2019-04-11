### Notes
Extract m4a from mp4: ffmpeg -i input.mp4 -vn -c:a copy output.m4a
Convert to wav: ffmpeg -i htf2.m4a htf2.wav 
Convert stereo wav to mono wav: ffmpeg -i input.wav -ac 1 output.wav

Remove intro silence: ffmpeg -i input.wav -af silenceremove=start_periods=1:start_threshold=-80dB output.wav
Trim: ffmpeg -i in.wav -af atrim=0:300 out.wav

#### Experimental
Not working tail silence remove
ffmpeg -i input1.wav -af silenceremove=start_periods=1:start_threshold=-80dB:stop_periods=1:stop_threshold=-80dB:stop_duration=0.1:detection=peak  output.wav 

### Motivation
To get rid of all those pesky intro and outros.

### How it should work
strip audio -> convert to wav -> make it mono -> strip silence -> trim to two minutes -> rename to intro
                                              -> reverse -> strip silence -> trim to two minutes -> rename to outro

                                            
### Shit hit the fan

tracks have to be aligned with a tolerance of 50 ms to get a 94% score.