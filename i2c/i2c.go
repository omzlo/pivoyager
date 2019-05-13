package i2c

/*
   This code largely borrows from i2c-dev.h, from:
       Copyright (C) 1995-97 Simon G. Vogl
       Copyright (C) 1998-99 Frodo Looijaard <frodol@dds.nl>

   It is released under the GPL like the original work:

   This program is free software; you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation; either version 2 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program; if not, write to the Free Software
   Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston,
   MA 02110-1301 USA.
*/

/*
#include <stdint.h>
#include <stdio.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/ioctl.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>

#define I2C_SMBUS_BLOCK_MAX 32  // As specified in SMBus standard
#define I2C_SMBUS_I2C_BLOCK_MAX 32  // Not specified but we use same structure
union i2c_smbus_data {
    uint8_t byte;
    uint16_t word;
    uint8_t block[I2C_SMBUS_BLOCK_MAX + 2]; // block[0] is used for length
    // and one more for PEC
};

// This is the structure as used in the I2C_SMBUS ioctl call
struct i2c_smbus_ioctl_data {
    char read_write;
    uint8_t command;
    int size;
    union i2c_smbus_data *data;
};

// smbus_access read or write markers
#define I2C_SMBUS_READ  1
#define I2C_SMBUS_WRITE 0

// SMBus transaction types (size parameter in the above functions)
//    Note: these no longer correspond to the (arbitrary) PIIX4 internal codes!
#define I2C_SMBUS_QUICK         0
#define I2C_SMBUS_BYTE          1
#define I2C_SMBUS_BYTE_DATA     2
#define I2C_SMBUS_WORD_DATA     3
#define I2C_SMBUS_PROC_CALL     4
#define I2C_SMBUS_BLOCK_DATA        5
#define I2C_SMBUS_I2C_BLOCK_BROKEN  6
#define I2C_SMBUS_BLOCK_PROC_CALL   7
#define I2C_SMBUS_I2C_BLOCK_DATA    8

// ioctl stuff
#define I2C_SLAVE   0x0703  // Change slave address
                // Attn.: Slave address is 7 or 10 bits
#define I2C_SMBUS   0x0720  // SMBus-level access



static int i2c_set_slave_address(int file, int address)
{
    if (ioctl(file, I2C_SLAVE, address) < 0) {
        fprintf(stderr,
                "Error: Could not set address to 0x%02x: %s\n",
                address, strerror(errno));
        return -errno;
    }
    return 0;
}

static int i2c_smbus_access(int file, char read_write, uint8_t command,
        int size, union i2c_smbus_data *data)
{
    struct i2c_smbus_ioctl_data args;

    args.read_write = read_write;
    args.command = command;
    args.size = size;
    args.data = data;
    return ioctl(file,I2C_SMBUS,&args);
}

//*******************************************************************

int i2c_open_dev(int bus)
{
    char filename[100];
    int file;

    snprintf(filename, 100, "/dev/i2c-%d", bus);
    filename[99]=0;
    file = open(filename, O_RDWR);

    return file;
}

int i2c_read_byte_data(int file, uint8_t addr, uint8_t reg)
{
    union i2c_smbus_data data;
    int res;

    if ((res = i2c_set_slave_address(file, addr))<0) {
        return res;
    }

    if (i2c_smbus_access(file, I2C_SMBUS_READ, reg, I2C_SMBUS_BYTE_DATA,&data))
        return -1;
    return 0x0FF & data.byte;
}

int i2c_write_byte_data(int file, uint8_t addr, uint8_t reg, uint8_t byte)
{
    union i2c_smbus_data data;
    int res;

    if ((res = i2c_set_slave_address(file, addr))<0) {
        return res;
    }

    data.byte = byte;

    if (i2c_smbus_access(file, I2C_SMBUS_WRITE, reg, I2C_SMBUS_BYTE_DATA,&data))
        return -1;
    return 0;
}

int i2c_read_block(int file, uint8_t addr, uint8_t reg, uint8_t *buf, uint8_t count)
{
    if (count>32 || count<1)
    {
        fprintf(stderr, "i2c block cannot be empty or more than 32 bytes.\n");
        return -1;
    }

    union i2c_smbus_data data;
    int res;

    if ((res = i2c_set_slave_address(file, addr))<0) {
        return res;
    }

    uint32_t size;
    if (count==32)
        size = I2C_SMBUS_I2C_BLOCK_BROKEN;
    else
        size = I2C_SMBUS_I2C_BLOCK_DATA;

    data.block[0] = count;

    if (i2c_smbus_access(file, I2C_SMBUS_READ, reg, size, &data)) {
        return -1;
    }
    for (int i=1;i<=data.block[0];i++)
        buf[i-1] = data.block[i];
    return data.block[0];
}

int i2c_write_block(int file, uint8_t addr, uint8_t reg, uint8_t *buf, uint8_t count)
{
    if (count>32 || count<1)
    {
        fprintf(stderr, "i2c block cannot be empty or more than 32 bytes.\n");
        return -1;
    }

    union i2c_smbus_data data;
    int res;

    if ((res = i2c_set_slave_address(file, addr))<0) {
        return res;
    }

    uint32_t size;
    if (count==32)
        size = I2C_SMBUS_I2C_BLOCK_BROKEN;
    else
        size = I2C_SMBUS_I2C_BLOCK_DATA;

    data.block[0] = count;
    for (int i=1;i<=count;i++)
        data.block[i] = buf[i-1];

    if (i2c_smbus_access(file, I2C_SMBUS_WRITE, reg, size, &data))
        return -1;
    return 0;
}

*/
// #cgo CFLAGS: -std=c99
import "C"

import "errors"

var (
	ReadError   = errors.New("I2C read error")
	WriteError  = errors.New("I2C write error")
	LengthError = errors.New("I2C buffer must be at most 32 bytes long")
)

type Bus C.int

func OpenBus(dev int) Bus {
	r := C.i2c_open_dev(C.int(dev))
	if r < 0 {
		return -1
	}
	return Bus(r)
}

func (i2c Bus) ReadByte(addr byte, reg byte) (byte, error) {
	r := C.i2c_read_byte_data(C.int(i2c), C.uchar(addr), C.uchar(reg))
	if r < 0 {
		return 0, ReadError
	}
	return byte(r), nil
}

func (i2c Bus) ReadBytes(addr byte, reg byte, data []byte) error {
	var buf [32]C.uchar
	if len(data) > 32 {
		return LengthError
	}
	r := C.i2c_read_block(C.int(i2c), C.uchar(addr), C.uchar(reg), &buf[0], C.uchar(len(data)))
	if r < 0 {
		return ReadError
	}
	for i := 0; i < len(data); i++ {
		data[i] = byte(buf[i])
	}
	return nil
}

func (i2c Bus) WriteByte(addr byte, reg byte, data byte) error {
	r := C.i2c_write_byte_data(C.int(i2c), C.uchar(addr), C.uchar(reg), C.uchar(data))
	if r < 0 {
		return WriteError
	}
	return nil
}

func (i2c Bus) WriteBytes(addr byte, reg byte, data []byte) error {
	var buf [32]C.uchar
	if len(data) > 32 {
		return LengthError
	}
	for i := 0; i < len(data); i++ {
		buf[i] = C.uchar(data[i])
	}
	r := C.i2c_write_block(C.int(i2c), C.uchar(addr), C.uchar(reg), &buf[0], C.uchar(len(data)))
	if r < 0 {
		return WriteError
	}
	return nil
}

func (i2c Bus) ModifyByte(addr byte, reg byte, mask byte, data byte) error {
	r, err := i2c.ReadByte(addr, reg)
	if err != nil {
		return err
	}
	w := (r & (^mask)) | (data & mask)
	return i2c.WriteByte(addr, reg, w)
}
