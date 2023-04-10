# Subscription API

## OpenAPI

I am going to be writing my docs by hand as an introduction to documenting apis. I think this will be fine for this particular project as there is only one endpoint. I am using [this](https://www.reddit.com/r/golang/comments/udfujj/any_good_openapi_3x_spec_generator_for_a_go_rest/) document as a base for learning how to define

## NATS

Interfacing with the CLI, it's possible to check info the NATS account, stream & consumers using some of the following commands:

```bash
# To view the current default account
$ nats context info
# To view the account and open connections
$ nats account report connections
# To view the streams
$ nats stream ls
# To view the consumers
$ nats consumers ls
```