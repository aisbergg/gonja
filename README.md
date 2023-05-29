<a name="readme-top"></a>

[![GoDoc](https://pkg.go.dev/badge/github.com/aisbergg/gonja)](https://pkg.go.dev/github.com/aisbergg/gonja/pkg/gonja)
[![GoReport](https://goreportcard.com/badge/github.com/aisbergg/gonja)](https://goreportcard.com/report/github.com/aisbergg/gonja)
[![Coverage Status](https://codecov.io/gh/aisbergg/gonja/branch/main/graph/badge.svg)](https://codecov.io/gh/aisbergg/gonja)
[![CodeQL](https://github.com/aisbergg/gonja/actions/workflows/codeql.yml/badge.svg
)](https://github.com/aisbergg/gonja/actions/workflows/codeql.yml)
[![License](https://img.shields.io/github/license/aisbergg/gonja)](https://pkg.go.dev/github.com/aisbergg/gonja)
[![LinkedIn](https://img.shields.io/badge/-LinkedIn-green.svg?logo=linkedin&colorB=555)](https://www.linkedin.com/in/andre-lehmann-97408221a/)

<br />
<br />
<div align="center">
  <a href="https://github.com/aisbergg/gonja">
    <img src="assets/logo.svg" alt="Logo" width="160" height="160">
  </a>

  <h2 align="center"><b>Gonja - Jinja2-like Templating</b></h2>

  <p align="center">
    Ignite your Go code with Jinja2-inspired magic and turbocharge your templating game!
    <br />
    <br />
    <a href="https://pkg.go.dev/github.com/aisbergg/gonja/pkg/gonja">View Docs</a>
    ·
    <a href="https://github.com/aisbergg/gonja/issues">Report Bug</a>
    ·
    <a href="https://github.com/aisbergg/gonja/issues">Request Feature</a>
  </p>
</div>

<details>
  <summary>Table of Contents</summary>

- [About](#about)
- [Installation](#installation)
- [Synopsis](#synopsis)
  - [General Usage](#general-usage)
  - [Custom Filters and Tests](#custom-filters-and-tests)
  - [Custom Value Types](#custom-value-types)
  - [Handling of Undefined Variables](#handling-of-undefined-variables)
  - [Extensions](#extensions)
  - [References](#references)
    - [Tests](#tests)
    - [Filters](#filters)
    - [Tags](#tags)
- [Roadmap](#roadmap)
- [Benchmark](#benchmark)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)

</details>



## About

Gonja is your go-to templating engine for Go, closely resembling the beloved Python counterpart, [Jinja2](https://jinja.palletsprojects.com/en/latest/). Seamlessly generating dynamic content for a myriad of use cases, including web applications, email personalization, report generation and many more. Say goodbye to complex coding and hello to a simple yet flexible solution for all your content generation needs.

**Features:**

- Jinja2-like templating engine for Go
- Supports template inheritance, macros, filters, tests, and more
- Customizable handling of undefined variables
- Usage of custom value types
- Extensibility by creating custom filters, tests and more

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Installation

```sh
go get github.com/aisbergg/gonja
```

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Synopsis

Gonja's template syntax is mostly identical to Jinja2. Some of the included filters and tests might vary. The [Jinja2 documentation](https://jinja.palletsprojects.com/en/latest/templates/) is a great source for information on how to write the templates.

### General Usage

- quick start
- customizing environment
    - You can find a [list of all options here](https://pkg.go.dev/github.com/aisbergg/gonja/pkg/gonja#Option).
- usage of tests and filters


### Custom Filters and Tests


### Custom Value Types


### Handling of Undefined Variables


### Extensions


### References

#### Tests

The following tests are included in Gonja:

| Name              | Description                                                       | Reference                                                                                         |
| ----------------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| `callable` | Return whether the object is callable (i.e., some kind of function). | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.callable)            |
| `defined` | Return true if the variable is defined. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#defined) |
| `divisibleby` | Return true if the variable is divisible by the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.divisibleby) |
| `eq`</br>`equalto`</br>`==` | Return true if the expression is equal to the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.eq) |
| `even` | Return true if the variable is even. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.even) |
| `ge`</br>`>=` | Return true if the expression is greater than or equal to the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.ge) |
| `gt`</br>`greaterthan`</br>`>` | Return true if the expression is greater than the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.gt) |
| `in` | Return true if the expression is contained in the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.in) |
| `iterable` | Return true if the variable is iterable. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.iterable) |
| `le`</br>`<=` | Return true if the expression is less than or equal to the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.le) |
| `lower` | Return a copy of the string with all the cased characters converted to lowercase. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.lower) |
| `lt`</br>`lessthan`</br>`<` | Return true if the expression is less than the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.lt) |
| `mapping` | Return true if the variable is a mapping (i.e., a dictionary). | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.mapping) |
| `ne`</br>`!=` | Return true if the expression is not equal to the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.ne) |
| `none` | Return true if the variable is None. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.none) |
| `number` | Return true if the variable is a number. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.number) |
| `odd` | Return true if the variable is odd. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.odd) |
| `sameas` | Return true if the expression is the same object as the argument. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.sameas) |
| `sequence` | Return true if the variable is a sequence (i.e., a list or tuple). | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.sequence) |
| `string` | Return true if the variable is a string. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.string) |
| `undefined` | Return true if the variable is undefined. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.undefined) |
| `upper` | Return a copy of the string with all the cased characters converted to uppercase. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-tests.upper) |

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



#### Filters

The following filters are included in Gonja:

| Name              | Description                                                       | Reference                                                                                         |
| ----------------- | ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| `abs`             | Return the absolute value of the argument.                        | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.abs)            |
| `attr`            | Get an attribute of an object dynamically.                        | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.attr)           |
| `batch`           | Group a sequence of objects into fixed-length chunks.             | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.batch)          |
| `bool`            | Convert the value to a boolean.                                   | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.bool)           |
| `boolean`         | Convert the value to a boolean.                                   | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.boolean)        |
| `capitalize`      | Capitalize the first character of a string.                       | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.capitalize)     |
| `center`          | Center a string in a field of a given width.                      | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.center)         |
| `default`</br>`d` | Return a default value if the value is undefined.                 | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.default)        |
| `dictsort`        | Sort a dictionary by key or value.                                | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.dictsort)       |
| `escape`</br>`e`  | Escape a string for HTML rendering.                               | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.escape)         |
| `filesizeformat`  | Convert a file size to a human-readable format.                   | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.filesizeformat) |
| `first`           | Get the first item of a sequence.                                 | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.first)          |
| `float`           | Convert the value to a floating-point number.                     | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.float)          |
| `forceescape`     | Escape a string for HTML rendering, even if it is marked as safe. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.forceescape)    |
| `format`          | Format a string using placeholders.                               | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.format)         |
| `groupby`         | Group a sequence of objects by a common attribute.                | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.groupby)        |
| `indent`          | Indent a string by a given number of spaces.                      | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.indent)         |
| `int`             | Convert the value to an integer.                                  | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.int)            |
| `integer`         | Convert the value to an integer.                                  | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.integer)        |
| `join`            | Join a sequence of strings with a delimiter.                      | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.join)           |
| `last`            | Get the last item of a sequence.                                  | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.last)           |
| `length`          | Get the length of a sequence or a string.                         | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.length)         |
| `list`            | Convert the value to a list.                                      | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.list)           |
| `lower`           | Convert a string to lowercase.                                    | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.lower)          |
| `map`             | Apply a filter to each item in a sequence.                        | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.map)            |
| `max`             | Get the maximum value in a sequence.                              | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.max)            |
| `min`             | Get the minimum value in a sequence.                              | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.min)            |
| `pprint`          | Pretty-print a Python object.                                     | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.pprint)         |
| `random`          | Get a random item from a sequence.                                | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.random)         |
| `reject`          | Remove items from a sequence that match a condition.              | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.reject)         |
| `rejectattr`      | Remove items from a sequence that have a certain attribute value. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.rejectattr)     |
| `replace`         | Replace occurrences of a substring with another string.           | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.replace)        |
| `reverse`         | Reverse the order of a sequence.                                  | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.reverse)        |
| `round`           | Round a number to a given number of decimal places.               | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.round)          |
| `safe`            | Mark a string as safe for HTML rendering.                         | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.safe)           |
| `select`          | Select items from a sequence that match a condition.              | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.select)         |
| `selectattr`      | Select items from a sequence that have a certain attribute value. | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.selectattr)     |
| `slice`           | Get a slice of a sequence.                                        | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.slice)          |
| `sort`            | Sort a sequence.                                                  | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.sort)           |
| `string`          | Convert the value to a string.                                    | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.string)         |
| `striptags`       | Remove HTML tags from a string.                                   | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.striptags)      |
| `sum`             | Get the sum of a sequence of numbers.                             | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.sum)            |
| `title`           | Convert a string to title case.                                   | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.title)          |
| `tojson`          | Convert a value to a JSON string.                                 | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.tojson)         |
| `trim`            | Remove whitespace from the beginning and end of a string.         | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.trim)           |
| `truncate`        | Truncate a string to a given length.                              | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.truncate)       |
| `unique`          | Remove duplicate items from a sequence.                           | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.unique)         |
| `upper`           | Convert a string to uppercase.                                    | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.upper)          |
| `urlencode`       | URL-encode a string.                                              | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.urlencode)      |
| `urlize`          | Convert URLs and email addresses in a string to clickable links.  | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.urlize)         |
| `wordcount`       | Count the number of words in a string.                            | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.wordcount)      |
| `wordwrap`        | Wrap a string to a given width.                                   | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.wordwrap)       |
| `xmlattr`         | Convert a dictionary to an XML attribute string.                  | [Jinja2 Ref](https://jinja.palletsprojects.com/en/latest/templates/#jinja-filters.xmlattr)        |

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



#### Tags



<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>

## Roadmap

- [ ] add comparable benchmarks; include Go built-in template engine, [pongo2](https://github.com/flosch/pongo2), [liquid](https://github.com/osteele/liquid)
- [ ] write more documentation
- [ ] write more tests
- [ ] clean up code
- [ ] optimize code and performance

## Benchmark

Inside the `benchmark` directory reside some comparable benchmarks that allow some performance comparison of gonja with other error handling libraries. The benchmarks can be executed by running `make bench`. Here are my results:

```plaintext
cpu: AMD Ryzen 5 5600X 6-Core Processor             
BenchmarkParse-12                   9858            107232 ns/op           85165 B/op        806 allocs/op
BenchmarkExecute-12                 7198            156427 ns/op           86365 B/op       1971 allocs/op
BenchmarkParallelExecute-12        16735             73162 ns/op          110357 B/op       1981 allocs/op
```

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Contributing

If you have any suggestions, want to file a bug report or want to contribute to this project in some other way, please read the [contribution guideline](CONTRIBUTING.md).

And don't forget to give this project a star 🌟! Thanks again!

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Contact

André Lehmann

- Email: aisberg@posteo.de
- [GitHub](https://github.com/aisbergg)
- [LinkedIn](https://www.linkedin.com/in/andre-lehmann-97408221a/)

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>



## Acknowledgments

Gonja, initially developed by [Axel Haustant](https://github.com/noirbizarre), was built on [_pongo2_](https://github.com/flosch/pongo2), a template engine inspired by Django, created by [Florian Schlachter](https://github.com/flosch). Many other awesome folks have also contributed to the code. Shoutout to all of you for doing an amazing job!

<p align="right"><a href="#readme-top" alt="abc"><b>back to top ⇧</b></a></p>
