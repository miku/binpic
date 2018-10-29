# binpic

Create a PNG from any file. Pitched by [c3er](https://github.com/c3er) at
[devopenspace](https://twitter.com/devopenspace) 2018 (see also:
[bin2img](https://github.com/c3er/bin2img)).

```shell
$ binpic /bin/ls
```

Encode a file as [grayscale image](https://golang.org/pkg/image/#Gray), optionally resize.

```shell
$ binpic -h
Usage of binpic:
  -o string
        output file, will be a PNG (default "output.png")
  -resize string
        resize, if set (default "0x0")
```

Thanks to the beautiful Go standard library packages like
[image](https://golang.org/pkg/image/) and [io](https://golang.org/pkg/io/),
this is little more complex than a *Hello World*. This [simple image
library](https://github.com/disintegration/imaging) helps, too.

## Install

This is a toy project, still want to try it out?

```shell
$ go get github.com/miku/binpic/cmd/...
```

## Gallery

### binpic binary (amd64)

![](output.png)

### binpic binary (arm)

![](gallery/arm.png)

### ls (coreutils)

![](gallery/ls.png)

### The go tool

![](gallery/go.png)

### Caffe model file

![](gallery/lenet.png)

### Protocol Buffer

![](gallery/pb.png)

### mp4

![](gallery/mp4.png)

### webm

![](gallery/webm.png)

### XML

![](gallery/xml.png)

### LevelDB file

![](gallery/ldb.png)

### PDF

![](gallery/pdf.png)

### MARC21

![](gallery/marc21.png)

### [DBM](https://en.wikipedia.org/wiki/Dbm)

![](gallery/dbm.png)

### wav

![](gallery/wav.png)

### d64

![](gallery/d64.png)

