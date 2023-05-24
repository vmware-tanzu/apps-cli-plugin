# CLI UX Component library

## Usage

To use the component package, import it into your Go program as follows:

``` go
import "github.com/vmware-tanzu/tanzu-plugin-runtime/component"
```

## AskForConfirmation Component

`AskForConfirmation` is a function that prompts the user to confirm or deny a choice with a message and returns an error if the response is not affirmative.

### Usage

``` go
package main

import (
    "github.com/pkg/errors"
    "github.com/vmware-tanzu/tanzu-plugin-runtime/component"
)

func main() {
    err := AskForConfirmation("Are you sure you want to delete this file?")
    if err != nil {
        return errors.Wrap(err, "failed to delete file")
    }

    // continue with file deletion
}
```

The function takes a message string as its argument and returns an error if the user's response is not affirmative (i.e., "y" or "Y"). If the response is affirmative, `AskForConfirmation` returns `nil`, and the program can continue executing.

If an error occurs during the function's execution, it will return a new error wrapped with the `github.com/pkg/errors` package.

## Colorable-TTY

`Colorable-TTY` is a Golang package that provides utilities for creating colorized and formatted console output. This package uses the `logrusorgru/aurora` package to generate the colorized output. It also includes several functions for formatting strings such as `Rpad`, `Underline`, `Bold`, `TrimRightSpace`, and `BeginsWith`.

Then, you can use the functions provided by this package. For example:

``` go
package main

import (
    "fmt"
    "github.com/vmware-tanzu/tanzu-plugin-runtime/component"
)

func main() {
    fmt.Println(component.Underline("This text is underlined!"))
    fmt.Println(component.Bold("This text is bold!"))
    fmt.Println(component.Rpad("Right padded text", 20))
    fmt.Println(component.TrimRightSpace("Trim the whitespace at the end   "))
    fmt.Println(component.BeginsWith("This is a test", "This"))
}
```

## Output Writer Component

This package provides a way to write output to different formats. It defines the `OutputWriter` interface which consists of `SetKeys`, `AddRow`, and `Render` methods.

### How to use

#### Initialization

``` go
func NewOutputWriter(output io.Writer, outputFormat string, headers ...string) OutputWriter
func NewObjectWriter(output io.Writer, outputFormat string, data interface{}) OutputWriter
```

- `NewOutputWriter` returns a new instance of `OutputWriter` and it accepts an `io.Writer`, an `outputFormat` (one of `TableOutputType`, `YAMLOutputType`, `JSONOutputType`, `ListTableOutputType`), and a variadic list of headers for the table.
- `NewObjectWriter` is the same as `NewOutputWriter` but it is used for writing objects instead of tables. It accepts an `io.Writer`, an `outputFormat` (one of `YAMLOutputType`, `JSONOutputType`), and an object that should be written.

#### Usage

``` go
func (ow *outputwriter) SetKeys(headerKeys ...string)
func (obw *objectwriter) SetKeys(headerKeys ...string)
```

`SetKeys` sets the values to use as the keys for the output values. It accepts a variadic list of strings.

``` go
func (ow *outputwriter) AddRow(items ...interface{})
func (obw *objectwriter) AddRow(items ...interface{})
```

`AddRow` appends a new row to our table. It accepts a variadic list of items. It is required to add items in the same order as the headers were provided.

``` go
func (ow *outputwriter) Render()
func (obw *objectwriter) Render()
```

`Render` emits the generated output to the output stream.

If `NewOutputWriter` was used for initialization, `Render` will generate output in the format specified by the `outputFormat` parameter.

If `NewObjectWriter` was used for initialization, `Render` will generate output in the format specified by the `outputFormat` parameter, and it will use the provided object for output.

#### Constants

``` go
const (
    TableOutputType OutputType = "table"
    YAMLOutputType OutputType = "yaml"
    JSONOutputType OutputType = "json"
    ListTableOutputType OutputType = "listtable"
)
```

- `TableOutputType` specifies that the output should be in table format.
- `YAMLOutputType` specifies that the output should be in yaml format.
- `JSONOutputType` specifies that the output should be in json format.
- `ListTableOutputType` specifies that the output should be in a list table format.

#### Example

``` go
package main

import (
    "fmt"
    "os"

    "github.com/vmware-tanzu/tanzu-plugin-runtime/component"
)

func main() {
    // Example usage of NewOutputWriter
    writer := output.NewOutputWriter(os.Stdout, output.TableOutputType, "Name", "Age")
    writer.AddRow("John", 30)
    writer.AddRow("Bob", 45)
    writer.Render()

    // Example usage of NewObjectWriter
    data := map[string]interface{}{
        "Name": "John",
        "Age": 30,
    }
    writer = output.NewObjectWriter(os.Stdout, output.JSONOutputType, data)
    writer.Render()
}
```

This will output:

``` go
+------+-----+
| Name | Age |
+------+-----+
| John | 30  |
| Bob  | 45  |
+------+-----+

{"Age":30,"Name":"John"}
```

## OutputWriterSpinner Component

`OutputWriterSpinner` is a Go package that provides an interface to `OutputWriter` augmented with a spinner. It allows for rendering output with a spinner while also providing the ability to stop the spinner and render the final output.

### Usage

To use `OutputWriterSpinner`, you can import the package in your Go code and create an instance of the `OutputWriterSpinner` interface using the `NewOutputWriterWithSpinner` function.

``` go
import "github.com/vmware-tanzu/tanzu-plugin-runtime/component"


// create new OutputWriterSpinner
outputWriterSpinner, err := component.NewOutputWriterWithSpinner(os.Stdout, "json", "Loading...", true)
if err != nil {
    fmt.Println("Error creating OutputWriterSpinner:", err)
    return
}

// Render output with spinner
outputWriterSpinner.RenderWithSpinner()

// Stop spinner and render final output
outputWriterSpinner.StopSpinner()
```

The `NewOutputWriterWithSpinner` function takes in the following parameters:

- `output io.Writer`: The output writer for the spinner and final output.
- `outputFormat string`: The output format for the final output. It can be either "json" or "yaml".
- `spinnerText string`: The text to display next to the spinner.
- `startSpinner bool`: Whether to start the spinner immediately or not.
- `headers ...string`: Optional headers for the final output.

The created `OutputWriterSpinner` instance provides two methods:

- `RenderWithSpinner()`: Renders the output with a spinner.
- `StopSpinner()`: Stops the spinner and renders the final output.

## Prompt Component

This is a Go package that provides a way to prompt the user for input in a CLI application.
It utilizes the `AlecAivazis/survey/v2` package to display a prompt to the user and capture their response.

Create a `PromptConfig` object to specify the prompt's `message`, `options`, `default` value, and `help` text. Then call the `Run` method with a pointer to a variable of the desired type to store the user's response.

``` go
var response string
p := &PromptConfig{
    Message: "Enter a value:",
    Default: "default value",
    Help: "Enter the value you would like to use.",
}
err := p.Run(&response)
if err != nil {
    // Handle error
}
fmt.Println("You entered:", response)
```

The `PromptConfig` struct also provides the Sensitive option to hide user input, and the `Options` field to provide a list of choices for the user to select from.

``` go
var response string
p := &PromptConfig{
    Message: "Enter your password:",
    Sensitive: true,
}
err := p.Run(&response)
if err != nil {
    // Handle error
}
fmt.Println("Your password is:", response)

p = &PromptConfig{
    Message: "Select an option:",
    Options: []string{"Option 1", "Option 2", "Option 3"},
}
err = p.Run(&response)
if err != nil {
    // Handle error
}
fmt.Println("You selected:", response)
```

### Example

``` go
package main

import (
    "fmt"
    "github.com/vmware-tanzu/tanzu-plugin-runtime/component"
)

func main() {
    var response string
    p := &prompt.PromptConfig{
        Message: "Enter your name:",
        Default: "John Doe",
        Help: "Enter your full name.",
    }
    err := p.Run(&response)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Hello,", response)
}
```

## Question Component

The `question.go` file provides a Go package component that implements prompting a CLI question and reading user response.

### Usage

#### Asking a Question

To prompt the user with a question, you can create a QuestionConfig object with the message to display and then call Run() method with a response object. For example:

``` go
package main

import (
    "fmt"

    "github.com/vmware-tanzu/tanzu-plugin-runtime/component"
)

func main() {
    q := component.QuestionConfig{
        Message: "What is your name?",
    }
    var name string
    err := q.Run(&name)
    if err != nil {
        // handle error
    }
    fmt.Printf("Hello, %s!\n", name)
}

```

The response object should be a pointer to a variable where the user's response will be stored.

This will display a prompt asking the user for their favorite color, with a default value of "blue" and a help message indicating the purpose of the question.

## Reader Component

The `reader.go` file provides a Go package component that implements reading input from a file or standard input (stdin).

To read input from a file, call ReadInput and pass the file path:

``` go
import "github.com/vmware-tanzu/tanzu-plugin-runtime/component"

func main() {
    filePath := "/path/to/file.txt"
    data, err := component.ReadInput(filePath)
    if err != nil {
        // Handle error
    }
    // Do something with the input data
    fmt.Println(string(data))
}
```

To read input from stdin, pass the - character as the file path to ReadInput:

``` go
import "github.com/vmware-tanzu/tanzu-plugin-runtime/component"

func main() {
    data, err := component.ReadInput("-")
    if err != nil {
        // Handle error
    }
    // Do something with the input data
    fmt.Println(string(data))
}
```

## Select Component

The `select.go` file provides a Go package component that implements a prompt for selecting an option. The package uses the survey library for prompting.

To prompt the user to select an option, create a SelectConfig object and pass it to the Select function:

``` go
import "github.com/vmware-tanzu/tanzu-plugin-runtime/component"

func main() {
    options := []string{"Option A", "Option B", "Option C"}
    config := &component.SelectConfig{
        Message: "Please select an option:",
        Default: "Option A",
        Options: options,
        Help: "Use arrow keys to move up and down, press Enter to select",
    }
    var response string
    err := component.Select(config, &response)
    if err != nil {
        // Handle error
    }
    // Do something with the selected option
    fmt.Println(response)
}
```
