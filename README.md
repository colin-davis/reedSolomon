# Reed-Solomon Error Correcting Algorithm

____

PROJECT STATE: `BETA`

This project has been tested for use with datamatrix decoding and performs accurately. If you find an error when using this for another application please create an issue with sufficient information or submit a PR.

---

## Install

With a [correctly configured](https://golang.org/doc/install#testing) Go toolchain:

`go get -u github.com/colin-davis/reedSolomon`

---
## Initializing Look Up Tables

Before using the decoder you will have to initialize the precomputed Galois Field look up tables that are used by the algorithm.
To do so you will need to provide the **Primative** and the **First Consecutive Root**. Every different system uses different values for this...
Below are some default values (If you find one that is not listed please let me know or submit and PR to the readme).

|            | Primative | FCR |
|------------|-----------|-----|
| QR-Codes   | 285       | 0   |
| Datamatrix | 301       | 1   |

## Examples

### Preamble
A Reed-Solomon encoded message is made up of two parts:
1. The encoded message symbols
2. The error correcting code symbols (ECC)

In the case of this package the message and ECC symbols are represented in a single
integer slice. The encoded message is at the start of the integer slice and the ECC symbols are appended to the end of the slice.


There are two types or erroneous symbols:
1. **Erasures** (the location of the missing symbol is known, we just don't know the correct value)
2. **Errors** (A symbol is incorrect, but we do not know which symbol is incorrect)

The algorithm can correct up to t erroneous symbols (t = the number of ECC symbols)
- Locating an unknown error cost 1
- Correcting an error cost 1

Therefore, the algorithm can correct up to t erasures because it does not need to locate the errors. Whereas the algorithm can only correct t/2 errors, because it has to locate the error first and then correct the error.

### Example 1: message with no errors
(The following examples use an ascii coded "hello world" message)
```go

import (
  "github.com/colin-davis/reedSolomon"
  "log"
)

func main() {

  // First we have to initialize the Galois Field Look Up Tables
  // This takes a "primitive" value that is different for each application...
  // For default QR-Codes use 285
  // For default Datamatrix use 301

  reedSolomon.InitGaloisFields(285, 0)

  // Get the Reed-Solomon encoded message as an int slice
  msg := []int{104, 101, 108, 108, 111,  32, 119, 111, 114, 108, 100, 145, 124, 96, 105, 94, 31, 179, 149, 163} // "hello world"
  numberEccSymbols := 9 // This message has 11 data symbols and 9 ECC symbols
  errorLocations := []int{}

  correctedMsg, correctedEcc, err := reedSolomon.Decode(msg, numberEccSymbols, errorLocations)
  if err != nil {
    log.Println(err)
  }

  log.Printf("\n Corrected MSG: %d \n Corrected ECC: %d", correctedMsg, correctedEcc)
}
```

### Example 2: message with 4 errors
```go

import (
  "github.com/colin-davis/reedSolomon"
  "log"
)

func main() {

  reedSolomon.InitGaloisFields(285, 0)

  msg := []int{104, 101, 108, 108, 111,  32, 119, 111, 114, 108, 100, 145, 124, 96, 105, 94, 31, 179, 149, 163} // "hello world"
  numberEccSymbols := 9
  errorLocations := []int{} // We don't provide the location of any errors (this will cost 2 for each error because it has to locate then fix the error)

  // Edit the msg to add 4 errors on purpose
  msg[3] = 11
  msg[6] = 92
  msg[14] = 2
  msg[17] = 42

  correctedMsg, correctedEcc, err := reedSolomon.Decode(msg, numberEccSymbols, errorLocations)
  if err != nil {
    log.Println(err)
  }

  log.Printf("\n Corrected MSG: %d \n Corrected ECC: %d", correctedMsg, correctedEcc)
}
```

### Example 3: message with 4 errors (but provide the location of 2 of the errors)
```go

import (
  "github.com/colin-davis/reedSolomon"
  "log"
)

func main() {

  reedSolomon.InitGaloisFields(285, 0)

  msg := []int{104, 101, 108, 108, 111,  32, 119, 111, 114, 108, 100, 145, 124, 96, 105, 94, 31, 179, 149, 163} // "hello world"
  numberEccSymbols := 9
  errorLocations := []int{3, 6} // We provide the location of the first two errors (These two errors will now only cost 1 each)

  // Edit the msg to add 4 errors on purpose
  msg[3] = 11
  msg[6] = 92
  msg[14] = 2
  msg[17] = 42

  correctedMsg, correctedEcc, err := reedSolomon.Decode(msg, numberEccSymbols, errorLocations)
  if err != nil {
    log.Println(err)
  }

  log.Printf("\n Corrected MSG: %d \n Corrected ECC: %d", correctedMsg, correctedEcc)
}
```

### Example 4: message with 5 errors (Too many to correct)
```go

import (
  "github.com/colin-davis/reedSolomon"
  "log"
)

func main() {

  reedSolomon.InitGaloisFields(285, 0)

  msg := []int{104, 101, 108, 108, 111,  32, 119, 111, 114, 108, 100, 145, 124, 96, 105, 94, 31, 179, 149, 163} // "hello world"
  numberEccSymbols := 9
  errorLocations := []int{0}

  // Edit the msg to add 5 errors on purpose
  msg[3] = 11
  msg[6] = 92
  msg[7] = 26
  msg[14] = 2
  msg[17] = 42

  correctedMsg, correctedEcc, err := reedSolomon.Decode(msg, numberEccSymbols, errorLocations)
  if err != nil {
    log.Println(err)
  }

  log.Printf("\n Corrected MSG: %d \n Corrected ECC: %d", correctedMsg, correctedEcc)
}
```

## Reed Solomon can be used for:

  - Datamatrix
  - Qr Codes

If you have used this for something not of the list let me know.


## Known issues:

  - First Consecutive Root (fcr) is currently hard coded to "1" (which works for datamatrix decoding but may need to be changed for other applications)

## TODO
 - Some code is still resembles the python origins and could be optimized or improved for GO.
 - Improve documentation on default primatives (prim) and first consecutive root (frc) for different applications (datamatrix, qr codes, etc...)
 - Allow selection of other algorithms (Berlekamp-Massey, Fourney, etc)

## Acknowledgments

This package was inspired from the python code at https://en.wikiversity.org/wiki/Reed%E2%80%93Solomon_codes_for_coders
