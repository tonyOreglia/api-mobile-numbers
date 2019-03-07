# API Mobile Numbers
A RESTful HTTP server which stores mobile numbers for later retrieval. 

The server will validate and attempt to fix mobile numbers before storing. 
Upon storing a batch of numbers, the service will return some basic information about the numbers stored.

## Usage

### Develop on Mac
This will auto-restart the server when changes are saved
```
$ brew install modd
$ modd
```

To run tests, 
```
$ go test ./...
```

**Adding Support for Additional Countries** 
Support for additional countries can be achieved by extending the `lookupRequirements` in `internal/server/fix.go`

### Run Server 
```
$ go run cmd/api-mobile-numbers/main.go 
```

### API

**Currently Supported Countries by International Olympic Committee (IOC) Code**
| Country | IOC Code | 
| ---- | ---- |
| South Africa | rsa |
| Australi | aus |
| Portugal | por |
| United States | usa |

REF [IOC Codes](https://en.wikipedia.org/wiki/List_of_IOC_country_codes)

#### Test a single number
 - Validate input number
 - Attempt to fix incorrectly formed number
 - Return if number is correct
 - Return correction details if number was corrected

```
POST http://localhost:80/<country_ioc_code>/numbers/test/<number>
```

For example, 
```
http://localhost:80/rsa/numbers/test/27640600114
```

**Response**
The response will have the content type application/json and will have the Format:
| Property | Type | Description | Required |
| ---- | ---- | ---- | ---- |
| valid | bool | indicates if the number provided is valid. If the number needed to be fixed, this will be False | Yes |
| number_provided | string | number in request parameter | No |
| number_fixed | string | number after being fixed | No |
| changes | string | comma separated list of changes | No |

#### Store CSV File of Numbers
```
POST http://localhost:80/<country_ioc_code>/numbers
```

Request Body CSV Format: 
| id | sms_phone | 
| --- | --- |
| 103343262 | 6478342944 | 
| 103426540 | 84528784843 |

Note that id is not used. 
Note that the header row cannot be ommited from the request body.

**Response Example**
```
{
    "ref": "3d836fe0-d2c8-4a79-adab-2f99f2b6ad88",
    "stats": {
        "valid_numbers_count": 463,
        "fixed_numbers_count": 533,
        "invalid_numbers_count": 4,
        "total_numbers_processed": 1000
    },
    "href": "http://localhost:80/numbers/3d836fe0-d2c8-4a79-adab-2f99f2b6ad88"
}
```


#### Return Details of Previously Processed File
Using the `ref` value from response above, call 
```
GET http://localhost:80/numbers/results/<ref>
```
to get the details of the previously processed file. Response is in the same format: 
```
{
    "ref": "3d836fe0-d2c8-4a79-adab-2f99f2b6ad88",
    "stats": {
        "valid_numbers_count": 463,
        "fixed_numbers_count": 533,
        "invalid_numbers_count": 4,
        "total_numbers_processed": 1000
    },
    "href": "http://localhost:80/numbers/3d836fe0-d2c8-4a79-adab-2f99f2b6ad88"
}
```

#### Download Previously Processed File 
Using the `href` value returned from a processed file, call 
```
GET http://localhost:80/numbers/3d836fe0-d2c8-4a79-adab-2f99f2b6ad88
```
for a JSON download of a previoulsy processed file. 

The response will have the content disposition attachment and will have the format
```
{
    "valid_numbers": [
        "27736529279",
        "27718159078",
        "27717278645",
    ],
    "fixed_numbers": [
        {
            "original_number": "730276061",
            "changes": "prepended number with 27,",
            "fixed_number": "27730276061"
        },
        {
            "original_number": "6478342944",
            "changes": "prepended number with 27,shortened number by removing 4,",
            "fixed_number": "27647834294"
        },
    ],
    "rejected_numbers": [
        "82192869",
        "2781441830",
        "8154255",
    ]
}
```

### Development Choices
Golang was chosen as an efficient backend language. 

Postgres was chosen as an effective way to store and retrieve data reliably.

 ### Corrections Made to Invalid Numbers
 1. If a number is too long, digits are trimmed from the end of the number
 2. If the number does not have the correct country dialing code, the dialing code is prepended 
 3. If there are any non-digits present, remove them

### Limitations 
 1. Largest file size possible
 2. 




