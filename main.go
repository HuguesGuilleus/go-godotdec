package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	outRoot := flag.String("out", "out", "Output root directory")
	flag.Parse()

	for _, info := range read(flag.Arg(0)).Files {
		log.Printf("file (%6d) [%d] %q", info.Size, info.Offset, info.Path)

		p := filepath.Join(*outRoot, filepath.FromSlash(info.Path))
		os.MkdirAll(filepath.Dir(p), 0o775)

		if err := os.WriteFile(p, info.Data, 0o666); err != nil {
			log.Fatal(err)
		}
	}
}

// Read package or fatal
func read(path string) (pkg *Package) {
	log.Println("open:", path)

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	pkg, err = ReadPackage(f)
	if err != nil {
		log.Fatal(err)
	}

	return
}
