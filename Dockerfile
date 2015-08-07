FROM golang:latest
ADD osd /bin/osd
CMD ["/bin/osd"]

