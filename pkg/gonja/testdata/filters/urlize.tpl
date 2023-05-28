{{ "http://www.example.org"|urlize|safe }}
{{ "http://www.example.org"|urlize(rel='nofollow')|safe }}
{{ "http://www.example.org"|urlize(rel='nofollow', target='_blank')|safe }}
{{ "http://www.example.org"|urlize(rel='noopener')|safe }}
{{ "www.example.org"|urlize|safe }}
{{ "example.org"|urlize|safe }}
--
{% filter urlize|safe %}
Please mail me at demo@example.com or visit mit on:
- lorem ipsum github.com/aisbergg/gonja lorem ipsum
- lorem ipsum http://www.example.org lorem ipsum
- lorem ipsum https://www.example.org lorem ipsum
- lorem ipsum https://www.example.org lorem ipsum
- lorem ipsum www.example.org lorem ipsum
- lorem ipsum www.example.org/test="test" lorem ipsum
{% endfilter %}
--
{% filter urlize(target='_blank', rel="nofollow")|safe %}
Please mail me at demo@example.com or visit mit on:
- lorem ipsum github.com/aisbergg/gonja lorem ipsum
- lorem ipsum http://www.example.org lorem ipsum
- lorem ipsum https://www.example.org lorem ipsum
- lorem ipsum https://www.example.org lorem ipsum
- lorem ipsum www.example.org lorem ipsum
- lorem ipsum www.example.org/test="test" lorem ipsum
{% endfilter %}
--
{% filter urlize(15)|safe %}
Please mail me at demo@example.com or visit mit on:
- lorem ipsum github.com/aisbergg/gonja lorem ipsum
- lorem ipsum http://www.example.org lorem ipsum
- lorem ipsum https://www.example.org lorem ipsum
- lorem ipsum https://www.example.org lorem ipsum
- lorem ipsum www.example.org lorem ipsum
- lorem ipsum www.example.org/test="test" lorem ipsum
{% endfilter %}
