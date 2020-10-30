# Starting a Go API
> [Go](https://golang.org/) is an open source programming language that makes it easy to build simple, reliable, and efficient software.

## Table of Contents
-   [Guide](#guide)
    -   [Install Go](#install-go)
    -   [Starting a project](#starting-a-project)
    -   [Creating Go package files](#creating-go-package-files)
    -   [Hello world HTTP server](#hello-world-http-server)
    -   [More control over the server](#more-control-over-the-server)
    -   [Adding tests](#adding-tests)
    -   [API contract using Protobuf](#api-contract-using-protobuf)
    -   [Make](#make)
    -   [Conversion endpoint test](#conversion-endpoint-test)
    -   [Adding celsius conversion endpoint](#adding-celsius-conversion-endpoint)
    -   [Supporting protobuf & json content types](#supporting-protobuf--json-content-types)
    -   [Next steps](#next-steps)
-   [Links](#links)
    -   [How to Write Go Code](#how-to-write-go-code)
    -   [Program Execution](#program-execution)
    -   [Code Testing](#code-testing)
    -   [Testing Package](#testing-package)
    -   [HTTP Package](#http-package)
    -   [Effective Go](#effective-go)
    -   [Go Fmt](#go-fmt)
    -   [Protocol Buffers](#protocol-buffers)
    -   [GNU Make](#gnu-make)
-   [Example Projects](#example-projects)
    -   [geolocation-api-golang](#geolocation-api-golang)
    -   [smc-config-api](#smc-config-api)
    -   [whatsmyip](#whatsmyip)
-   [Misc](#misc)
    -   [Visual Studio Code test coverage gutters](#visual-studio-code-test-coverage-gutters)
-   [Maintenance](#maintenance)

## Guide
This is an intro to writing a go API starting from scratch, building a simple "hello world" server then finally building it into an API which converts Fahrenheit to Celsius.  
The code you would have at the end of following this guide is available in this repo [here](./tempconvert)

### Install Go
-   You can download Go for Windows, MacOS, or Linux [here](https://golang.org/dl/)
    -   On MacOS you can use `brew install golang`

### Starting a project
1.  Create our project directory and init our go module
    -   Further reading [here](#how-to-write-go-code)
    ```sh
    mkdir tempconvert
    cd tempconvert
    go mod init example.com/user/tempconvert
    # Open directory in your IDE
    ```

### Creating Go package files
1.  Create our main package file `main.go`
    -   The main package must have package name `main` and declare a function `main()`
    -   Further reading [here](#program-execution)

    `main.go`
    ```go
    package main

    func main() {

    }
    ```

### Hello world HTTP server
1.  Update `main()` within `main.go`
    -   Further reading [here](#http-package)
    -   This guide assumes your IDE automatically adds imports on save
    ```go
    func main() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("Hello World!"))
        })

        log.Print("Listening")
        log.Fatal(http.ListenAndServe(":8080", nil))
    }
    ```
2.  Run the server
    ```sh
    go run main.go
    ```
3.  In your browser visit <http://localhost:8080>, if all goes well you should see "Hello World!"

### More control over the server
1.  Create a function in `main.go` to handle requests to `/`
    ```go
    func helloHandler(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World!"))
    }
    ```
2.  Create a function in `main.go` that creates & returns our API's [multiplexer](https://golang.org/pkg/net/http/#ServeMux)
    -   Functions with a name that begins with an upper case letter are exported, other functions are not.
    -   Functions that are exported must have a comment above them that begins with their name
    -   We've exported it for use in `main_test.go` later
    -   Further reading [here](#effective-go)
    ```go
    // GetMux returns the multiplexer - registered routes & functions
    func GetMux() http.Handler {
        mux := http.NewServeMux()
        mux.HandleFunc("/", helloHandler)
        return mux
    }
    ```
3.  Update `main()` to define our server settings and start the server
    -   This guide assumes your IDE automatically formats using `gofmt` on save.
    -   Further reading on `gofmt` [here](#go-fmt)
    -   Further reading on `http.Server` [here](#http-package)
    ```go
    func main() {
        s := &http.Server{
            Handler:      GetMux(),
            Addr:         ":8080",
            ReadTimeout:  10 * time.Second,
            WriteTimeout: 10 * time.Second,
            IdleTimeout:  1 * time.Minute,
        }
        log.Fatal(s.ListenAndServe())
    }
    ```
4.  Run the server and check that the API responds with "Hello World!" in your browser
    ```sh
    go run main.go
    ```
    <http://localhost:8080>

### Adding tests
At this point it is time to write the actual logic of our API and having tests will allow us to employ Test Driven Development (TDD) and keeps our code maintainable.

1.  Create a tests file used to test `main.go`
    -   Tests are written inside files that end in `_test.go`
    -   Test functions are named like `TestXxx`
    -   Further reading [here](#code-testing) and [here](#testing-package)

    `main_test.go`
    ```go
    package main

    import (
        "os"
        "testing"
    )

    var mux http.Handler

    // This is a special function used to run code before and after testing runs
    func TestMain(m *testing.M) {
        // Code here runs before testing starts
        mux = GetMux()
        // Run tests
        exitCode := m.Run()
        // Code here runs after testing finishes
        os.Exit(exitCode)
    }
    ```
2.  To keep the tests DRY, in `main_test.go` define a function which can be used to make a mock request using the mux handler
    ```go
    func mockRequest(method string, url string) ([]byte, *http.Response, error) {
        request := httptest.NewRequest(method, url, nil)
        recorder := httptest.NewRecorder()
        mux.ServeHTTP(recorder, request)
        resp := recorder.Result()
        body, err := ioutil.ReadAll(recorder.Body)
        return body, resp, err
    }
    ```
3.  Write a test which tests the API response to a request which currently responds with "Hello World!"
    ```go
    func TestHelloOk(t *testing.T) {
        body, resp, err := mockRequest("GET", "http://localhost:1234/")
        require.NoError(t, err)
        require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
        require.Equal(t, "Hello World!", string(body))
    }
    ```
4.  Run the test
    ```sh
    go test -v -coverprofile=cover.out ./...
    # === RUN   TestHelloOk
    # --- PASS: TestHelloOk (0.00s)
    # PASS
    # coverage: 66.7% of statements
    # ...
    ```

### API contract using Protobuf
Defining an API contract using a `.proto` file for Go makes our code robust by naturally providing our code with correctly typed request and reply objects. This is beneficial even if we do not plan on only responding with the `application/protobuf` content type.
Further reading [here](#protocol-buffers)

1.  Create a new folder in the root directory of your project (`tempconvert/contract`)
2.  Within `contract` create a new file named `contract.proto` where we will write our API contract

    `contract.proto`
    ```proto
    syntax = "proto3";

    option go_package = "tempconvert/contract";

    package contract;

    message TempConvertRequest { double fahrenheit = 1; }
    message TempConvertReply { double celsius = 1; }
    ```
3.  To compile it you'll need to have protobuf installed
    ```sh
    go get github.com/golang/protobuf/protoc-gen-go
    ```
4.  Compile the `.proto` file into a `.go` file
    ```sh
    protoc contract/contract.proto --go_out=contract
    mv contract/tempconvert/contract/* contract
    rm -rf contract/tempconvert
    ```
    -   This should generate a file `contract/contract.pb.go`
    -   If you run into "command not found" errors
        ```sh
        export PATH=$PATH:$HOME/go/bin
        export PATH=$PATH:/usr/local/go/bin
        ```

### Make
We are having to write a lot of commands in the terminal so a tool like `make` can help save time. Among other things, a `Makefile` is a decent way of providing single command aliases for various commands. Further reading [here](#make-docs)

1.  Create a Makefile in the root directory of your repo

    `Makefile`
    ```makefile
        protoc:
            rm -f contract/*.pb.go
            protoc contract/contract.proto --go_out=contract
            mv contract/tempconvert/contract/* contract
            rm -rf contract/tempconvert
        .PHONY: protoc

        test:
            go test -v -coverprofile=cover.out ./...
        .PHONY: test

        run:
            go run main.go
        .PHONY: run
    ```
2.  Try running each of the commands:
    ```sh
    make protoc # Recompiles the API contract
    make test   # Runs tests
    make run    # Run the API
    ```

### Conversion endpoint test
1.  Add a new test for the endpoint we want to add within `main_test.go`
    ```go
    func TestConvertOk(t *testing.T) {
        body, resp, err := mockRequest("GET", "http://localhost:1234/celsius?fahrenheit=32")
        require.NoError(t, err)
        require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
        require.Equal(t, "0", string(body))
    }
    ```
2.  Run the tests again, they will fail because we haven't added any of what this is testing but that's okay, in the next portion of the guide we will add it.
    ```sh
    make test
    # go test -v -coverprofile=cover.out ./...
    # === RUN   TestHelloOk
    # --- PASS: TestHelloOk (0.00s)
    # === RUN   TestConvertOk
    #     main_test.go:45: 
    #                 Error Trace:    main_test.go:45
    #                 Error:          Not equal: 
    #                                 expected: "0"
    #                                 actual  : "Hello World!"
    #                                 ...
    #                 Test:           TestConvertOk
    # --- FAIL: TestConvertOk (0.00s)
    ```
    The API is still responding based on our root route `/` which says hello.

### Adding celsius conversion endpoint
1.  In `main.go` add a new function which will handle our conversion requests
    -   You'll note that the value is hardcoded, following TDD this is okay, we will fix it in the next TDD cycle
    ```go
    func celsiusHandler(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("0"))
    }
    ```
2.  In `main.go` update the `GetMux` function to register the new function as a route
    ```go
    // GetMux returns the multiplexer - routes & registered functions
    func GetMux() http.Handler {
        mux := http.NewServeMux()
        mux.HandleFunc("/", helloHandler)
        mux.HandleFunc("/celsius", celsiusHandler)
        return mux
    }
    ```
3.  Rerun tests now that the new endpoint has been added
    ```sh
    make test
    # ...
    # === RUN   TestHelloOk
    # --- PASS: TestHelloOk (0.00s)
    # === RUN   TestConvertOk
    # --- PASS: TestConvertOk (0.00s)
    # PASS
    # coverage: 75.0% of statements
    # ...
    ```
4.  If we change the test values it will fail because we hardcoded the response to be "0"
5.  So add conversion code in `main.go` in a new resolver function
    ```go
    func celsiusResolver(r *contract.TempConvertRequest) *contract.TempConvertReply {
        c := (r.Fahrenheit - 32) * 5 / 9
        return &contract.TempConvertReply{Celsius: c}
    }
    ```
6.  Update `celsiusHandler` to use the query param and call the resolver function
    ```go
    func celsiusHandler(w http.ResponseWriter, r *http.Request) {
        fahrenheit, ok := r.URL.Query()["fahrenheit"]
        if !ok {
            http.Error(w, "missing fahrenheit URL query param", http.StatusBadRequest)
            return
        }
        f, err := strconv.ParseFloat(fahrenheit[0], 64)
        if err != nil {
            http.Error(w, "invalid fahrenheit value", http.StatusBadRequest)
            return
        }
        reply := celsiusResolver(&contract.TempConvertRequest{Fahrenheit: f})
        w.Write([]byte(strconv.FormatFloat(reply.Celsius, 'g', -1, 64)))
    }
    ```
7.  Run the tests again
    ```sh
    make test
    # ...
    # === RUN   TestHelloOk
    # --- PASS: TestHelloOk (0.00s)
    # === RUN   TestConvertOk
    # --- PASS: TestConvertOk (0.00s)
    # PASS
    # coverage: 68.4% of statements
    # ...
    ```
8.  Try it out
    -   Terminal 1
        ```sh
        make run
        # go run main.go
        # YYYY/MM/DD HH/MM/SS Listening
        ```
    -   Terminal 2
        ```sh
        curl "localhost:8080/celsius?fahrenheit=32"
        # 0
        curl "localhost:8080/celsius?fahrenheit=108"
        # 42.22222222222222
        curl "localhost:8080/celsius?fahrenheit=20"
        # -6.666666666666667
        ```

### Supporting protobuf & json content types
1. Add a marhalling function to `main.go`
    ```go
    func celsiusMarshaller(w http.ResponseWriter, r *http.Request, reply *contract.TempConvertReply) ([]byte, error) {
        accept := r.Header.Get("accept")
        w.Header().Set("Content-Type", accept)
        switch accept {
        case "application/protobuf":
            return proto.Marshal(reply)
        case "application/json":
            return json.Marshal(reply)
        default:
            w.Header().Set("Content-Type", "text/plain")
            return []byte(strconv.FormatFloat(reply.Celsius, 'g', -1, 64)), nil
        }
    }
    ```
2. Update the handler in `main.go` to make use of the marshaller
    ```go
    func celsiusHandler(w http.ResponseWriter, r *http.Request) {
        fahrenheit, ok := r.URL.Query()["fahrenheit"]
        if !ok {
            http.Error(w, "missing fahrenheit URL query param", http.StatusBadRequest)
            return
        }
        f, err := strconv.ParseFloat(fahrenheit[0], 64)
        if err != nil {
            http.Error(w, "invalid fahrenheit value", http.StatusBadRequest)
            return
        }
        reply := celsiusResolver(&contract.TempConvertRequest{Fahrenheit: f})
        body, err := celsiusMarshaller(w, r, reply)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Write(body)
    }
    ```
4. Run  the server and test it out
    -   Terminal 1
        ```sh
        make run
        # go run main.go
        # YYYY/MM/DD HH/MM/SS Listening
        ```
    -   Terminal 2
        ```sh
        curl -H "accept: application/json" "http://localhost:8080/celsius?fahrenheit=68"
        # {"celsius":20}
        ```

### Next steps
Some example next steps not covered in this guide
-   Add tests for each content type
-   Add [benchmarking](#testing-package)
-   Run your tests in a github action
    -   See projects in the [example projects](#example-projects) section which use `proto.Marshal` and `json.Marshal` to do so.
-   Refactor endpoint to function dynamically like `/tempTo?tempFrom=tempFromValue` so that you could use...
    -   `/fahrenheit?celsius=30`
    -   `/celsius?kelvin=0`
    -   etc...

## Links
Links and resources

### How to Write Go Code
<https://golang.org/doc/code.html>

### Program Execution
<https://golang.org/ref/spec#Program_execution>

### Code Testing
<https://golang.org/doc/code.html#Testing>

### Testing Package
<https://golang.org/pkg/testing/>
<https://golang.org/pkg/testing/#hdr-Benchmarks>

### HTTP Package
<https://golang.org/pkg/net/http/>
<https://golang.org/pkg/net/http/#Server>

### Effective Go
<https://golang.org/doc/effective_go.html#names>

### Go Fmt
<https://golang.org/pkg/fmt/>

### Protocol Buffers
<https://developers.google.com/protocol-buffers>

### GNU Make
<https://www.gnu.org/software/make/manual/make.html>

## Example Projects

### geolocation-api-golang
<https://github.com/searchspring/geolocation-api-golang>

### smc-config-api
<https://github.com/searchspring/smc-config-api>

### whatsmyip
<https://github.com/wafer-bw/whatsmyip>

## Misc
Other useful information

### Visual Studio Code test coverage gutters
1.  `cmd/ctrl+shift+p`
2.  Type "preferences"
3.  Select `Preferences: Open Settings (JSON)`
4.  Copy and paste the below & save
        "go.coverOnSave": true,
        "go.coverageDecorator": {
            "type": "gutter",
            "coveredHighlightColor": "rgba(64,128,128,0.5)",
            "uncoveredHighlightColor": "rgba(128,64,64,0.25)",
            "coveredGutterStyle": "blockgreen",
            "uncoveredGutterStyle": "blockred"
        },
        "go.coverOnSingleTest": true,
5.  Resave `main.go` and it should show red & green gutters to the left of line the line numbers for each uncovered & covered line

## Maintenance
Updating this README's Table of Contents:
1.  Save it as `README.md`  
2.  Install and run `remark` using `remark-toc`
    ```sh
    npm i -g remark
    npm i -g remark-toc
    remark README.md --use toc --output
    ```
