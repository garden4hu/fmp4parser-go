# fmp4parser-go
fpm4parser is a mp4 parser. It supports parsing most of the boxes specified in ISO/IEC 14496-12. In addition, it also supports the parsing of encryption boxes specified in ISO/IEC 14496-7.

The purpose of fmp4parser-go is to provide as complete an overview as possible for MP4, rather than simply listing the contents of the boxes.  It will also try to provide some basic interfaces for using. 

**Status: Under construction**

    Note: Because this tool is in a personal maintenance state, it lacks comprehensive and extensive testing.
        If you need a reliable and proven library to use for your production projects, please consider FFMPEG. 
        Of course, this project is still not available at this stage.

fmp4parser implements the parsing of the following boxes:

| Type |  |  |  |  |  | Remark |
|---|---|---|---|---|---|---|
| moov |  |  |  |  |  |  |
|  | mvhd |  |  |  |  |  |
|  | trak |  |  |  |  |  |
|  |  | tkhd |  |  |  |  |
|  |  | tref |  |  |  |  |
|  |  | trgr |  |  |  |  |
|  |  | edts |  |  |  |  |
|  |  |  | elst |  |  |  |
|  |  | senc |  |  |  |  |
|  |  | mdia |  |  |  |  |
|  |  |  | mdhd |  |  |  |
|  |  |  | hdlr |  |  |  |
|  |  |  | elng |  |  |  |
|  |  |  | minf |  |  |  |
|  |  |  |  | vmhd |  |  |
|  |  |  |  | smhd |  |  |
|  |  |  |  | sthd |  |  |
|  |  |  |  | dinf |  |  |
|  |  |  |  |  | dref |  |
|  |  |  |  | stbl |  |  |
|  |  |  |  |  | stsd |  |
|  |  |  |  |  | stts |  |
|  |  |  |  |  | ctts |  |
|  |  |  |  |  | cslg |  |
|  |  |  |  |  | stsc |  |
|  |  |  |  |  | stsz |  |
|  |  |  |  |  | stz2 |  |
|  |  |  |  |  | stco |  |
|  |  |  |  |  | co64 |  |
|  |  |  |  |  | stss |  |
|  |  |  |  |  | stsh |  |
|  |  |  |  |  | padb |  |
|  |  |  |  |  | stdp |  |
|  |  |  |  |  | sdtp |  |
|  |  |  |  |  | sbgp |  |
|  |  |  |  |  | sgpd |  |
|  |  |  |  |  | subs |  |
|  |  |  |  |  | saiz |  |
|  |  |  |  |  | saio |  |
|  |  |  |  |  | senc |  |
|  | pssh |  |  |  |  |  |
|  | mvex |  |  |  |  |  |
|  |  | mehd |  |  |  |  |
|  |  | trex |  |  |  |  |
|  |  | leva |  |  |  |  |
| moof |  |  |  |  |  |  |
|  | mfhd |  |  |  |  |  |
|  | traf |  |  |  |  |  |
|  |  | tfhd |  |  |  |  |
|  |  | trun |  |  |  |  |
|  |  | tfdt |  |  |  |  |
|  |  | sbgp |  |  |  |  |
|  |  | sgpd |  |  |  |  |
|  |  | subs |  |  |  |  |
|  |  | saiz |  |  |  |  |
|  |  | saio |  |  |  |  |
|  | pssh |  |  |  |  |  |
| mdat |  |  |  |  |  |  |
| free |  |  |  |  |  |  |
| skip |  |  |  |  |  |  |
| styp |  |  |  |  |  |  |
| sidx |  |  |  |  |  |  |
| ssix |  |  |  |  |  |  |
