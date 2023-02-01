## Loggie test

Loggie test is a project for automated testing of Loggie, mainly including:

- e2e test: end-to-end test from the consumption (collection) source to the functional test sent to the sink, mainly used to verify the correctness of the function
- Performance test: including regularly obtaining Loggie's indicators at this time and generating charts
- Chaos test: test and verification of some abnormal scenarios, such as frequent reload, whether it will cause abnormal data sent by Loggie


#### Code structure

- /e2e: end-to-end test
- /performance: performance test
- /chaos: chaos test
- /pkg
    - /env: abstractly represents the environment information of the test, the environment will initialize the following dependent resources
    - /resources: Represents resources, including externally dependent services, such as Elasticsearch, etc.
    - /cfg: The configuration structure of each test set, mainly including resources and cases
    - /report: metrics graph output

#### Contributing

1. Identify the classification of the test set to be written (e2e/performance/chaos), and then refer to the existing code in the corresponding directory structure, and add or modify it.
2. The test framework uses [ginkgo](https://github.com/onsi/ginkgo) and [gomega](https://github.com/onsi/gomega). Please understand the basic usage of the framework first.