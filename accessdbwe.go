// Copyright 2017 Benjamin 'Benno' Falkner. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

// code is based on some c experiments (more efficient less platform independent)
// feel free to implement this code in CGO

/** mdb.c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MDB_PAGE_SIZE 4096

#define MDB_VER_JET3 0
#define MDB_VER_JET4 1
#define MDB_VER_ACCDB2007 0x02
#define MDB_VER_ACCDB2010 0x0103


static char  JET3_XOR[] = { 0x86,0xfb,0xec,0x37,0x5d,0x44,0x9c,0xfa,0xc6,
                            0x5e,0x28,0xe6,0x13,0xb6,0x8a,0x60,0x54,0x94};

static short JET4_XOR[] = { 0x6aba,0x37ec,0xd561,0xfa9c,0xcffa,
                            0xe628,0x272f,0x608a,0x0568,0x367b,
                            0xe3c9,0xb1df,0x654b,0x4313,0x3ef3,
                            0x33b1,0xf008,0x5b79,0x24ae,0x2a7c};

size_t readMDBPage(char *filen, size_t pageSize, char *buffer)
{
    size_t r;
    FILE *f;

    f = fopen(filen,"rb");
    if( f==NULL ) {
        fprintf(stderr,"ERROR: could not open %s\n",filen);
        exit(1);
    }
    r = fread(buffer,1,pageSize,f);
    if( r != pageSize ) {
        fprintf(stderr,"ERROR: could not read page from %s (%li/%li)\n",filen,r,pageSize);
        exit(1);
    }
    fclose(f);
    return r;
}

int scanMDBPage(char *buffer){
    int i;
    int version;
    char pwd[40];
    short *pwd4 = (short*)pwd;
    short magic;

    //page id
    if(buffer[0]!=0){
        fprintf(stderr,"ERROR no vaild db\n");
        return 1;
    }

    //Version
    version = *((int*)&buffer[0x14]);
    switch(version){
        case MDB_VER_JET3:
                printf("DB Version: JET 3\n");
                break;
        case MDB_VER_JET4:
                printf("DB Version: JET 4\n");
                break;
        case MDB_VER_ACCDB2007:
                printf("DB Version: AccessDB 2007\n");
                break;
        case MDB_VER_ACCDB2010:
                printf("DB Version: AccessDB 2010\n");
                break;
        default:
                fprintf(stderr,"ERROR unkown version: %x\n",version);
                return 1;
    }


    //Password extract
    if( version==0 ) { // JET 3
        memcpy(pwd,buffer+0x42,20);
        for(i=0;i<18;i++){
            pwd[i] ^= JET3_XOR[i];
        }
        printf("Password: %20s\n",pwd);

    } else if ( version==1 ) { // JET 4
        memcpy(pwd,buffer+0x42,40);
        magic = *((short*)&buffer[0x66]);
        magic ^= JET4_XOR[18];

        for(i=0;i<18;i++){
            pwd4[i] ^= JET4_XOR[i];
            if(pwd4[i]>255){
                pwd4[i] ^= magic;
            }
            pwd[i] = pwd4[i];
        }
	        printf("Password: %20s\n",pwd);
    }
    return 0;
}

int main (int argc, char **argv)
{
    char page[MDB_PAGE_SIZE];

    //Check Arguments
    printf("MDB Access Tool\n");
    if (argc < 2){
        printf("Missing: file\n");
        return 1;
    }

    //Read data
    printf("Reading: %s\n", argv[1]);
    readMDBPage(argv[1],MDB_PAGE_SIZE,page);
    scanMDBPage(page);

    return 0;
}
**/

package accessdbwe

import (
	"database/sql"
	"encoding/binary"
	"errors"
	_ "github.com/mattn/go-adodb"
	"io"
	"os"
	"strings"
)

//
// Implementing decoding of the so called password
//

// setting XOR Values
var JET3_XOR = []byte{0x86, 0xfb, 0xec, 0x37, 0x5d, 0x44, 0x9c, 0xfa, 0xc6,
	0x5e, 0x28, 0xe6, 0x13, 0xb6, 0x8a, 0x60, 0x54, 0x94}
var JET4_XOR = []byte{0x6a, 0xba, 0x37, 0xec, 0xd5, 0x61, 0xfa, 0x9c, 0xcf, 0xfa,
	0xe6, 0x28, 0x27, 0x2f, 0x60, 0x8a, 0x05, 0x68, 0x36, 0x7b,
	0xe3, 0xc9, 0xb1, 0xdf, 0x65, 0x4b, 0x43, 0x13, 0x3e, 0xf3,
	0x33, 0xb1, 0xf0, 0x08, 0x5b, 0x79, 0x24, 0xae, 0x2a, 0x7c}

func readEncoding(filen string) (string, error) {
	file, err := os.Open(filen)
	if err != nil {
		return "", err
	}
	defer file.Close()

	p, err := readPage(file)
	if err != nil {
		return "", err
	}

	if p[0] != 0 {
		return "", errors.New("ERROR no vaild db")
	}

	buf := make([]byte, 20)
	switch binary.LittleEndian.Uint32(p[0x14:0x18]) {
	case 0x0: //Jet3
		for i := 0; i < 18; i++ {
			buf[i] = p[0x42+i] ^ JET3_XOR[i]
		}

	case 0x1, 0x02, 0x0103: //Jet4
		// sorry messed up byteorder
		p[0x66] = p[0x66] ^ JET4_XOR[36+1]
		p[0x67] = p[0x67] ^ JET4_XOR[36]
		for i := 0; i < 18; i++ {
			p[0x42+2*i] = p[0x42+2*i] ^ JET4_XOR[2*i+1]
			p[0x42+2*i+1] = p[0x42+2*i+1] ^ JET4_XOR[2*i]

			if p[0x42+2*i+1] > 0 {
				p[0x42+2*i] = p[0x42+2*i] ^ p[0x66]
				p[0x42+2*i+1] = p[0x42+2*i+1] ^ p[0x67]
			}
			buf[i] = p[0x42+2*i]
		}
	//case 0x02:    //ACCDB2007  //should use same encoding as JET4
	//case 0x0103:  //ACCDB2010
	default:
		return "", errors.New("ERROR unknown encoding")
	}

	var i int
	for i = 0; i < 20; i++ {
		if buf[i] == 0 {
			break
		}
	}
	return string(buf[:i]), nil

}

func readPage(r io.Reader) ([]byte, error) {
	p := make([]byte, 4096)
	s, err := r.Read(p)
	if err != nil {
		return nil, err
	}
	if s != 4096 {
		return p, errors.New("ERROR incomplete page")
	}
	return p, nil
}

//
// Implementing new open function
//
func Open(driver, filen string) (*sql.DB, error) {
	var err error
	var db *sql.DB

	if driver == "adodb" {
		// extract filename from filen string
		toks := strings.Split(filen, ";")
		for _, tok := range toks {
			pair := strings.Split(tok, "=")
			if len(pair) > 1 && pair[0] == "Data Source" {
				enc, err := readEncoding(pair[1])
				if err != nil {
					return nil, err
				}
				filen = filen + `;Jet OLEDB:Database Password=` + enc + `;`
			}
		}
	}

	db, err = sql.Open(driver, filen)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
