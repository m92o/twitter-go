DIRS=\
	src/pkg\
	src/cmd

all.dirs: $(addsuffix .all, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
clean.dirs: $(addsuffix .clean, $(DIRS))

%.all:
	cd $* && make all

%.install:
	cd $* && make install

%.clean:
	cd $* && make clean

all: all.dirs

install: install.dirs

clean: clean.dirs

