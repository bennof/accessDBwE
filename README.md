# accessDBwE

Microsoft ADODB driver conforming to the built-in database/sql interface with automated encoding

This is based on go-adodb from Yasuhiro Matsumoto (a.k.a mattn)

## Installation 

This should work:

    go get github.com/bennof/accessDBwE

## Documantation 

For API and usage see:

    http://godoc.org/github.com/mattn/go-adodb

The only addition is: reading out the so called password from the mdb or accessdb file. 

```go
import (
    ...
    "database/sql"
    "accessDBwE"
)

func main()  {
    var db *sql.DB
    var err error 
    db, err = accessDBwE.Open("adodb","Provider=Microsoft.ACE.OLEDB.12.0;Data Source=SomeFile.mdb;")

    // use db like any other sql.db handle
}
```

The "Jet OLEDB:Database Password=" key, value pair will be added automatically if it is needed.

Be carefull setting the provider to correct drivers.  

## License

BSD (3-Clause)

## Author

Benjamin Benno Falkner 
