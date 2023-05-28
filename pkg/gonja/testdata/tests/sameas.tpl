{{ simple is sameas simple }}  # map is not addressable
{{ simple is sameas(simple) }}  # same output, parentheses for arg
{{ 123456 is sameas simple }}
{{ 123456 is sameas 123456 }}
{{ number is sameas 11 }}
{{ simple.nil is sameas nil }}
{{ simple.nil is sameas None }}
{{ complex.user.Name is sameas complex.user.Name }}
