# **apimock** language specification

## Purpose

The **apimock** format is intended to be an **executable contract** for HTTP services.  It allows you to concisely describe both the **request** and the **possible responses** of a route, so that a mock server can emulate the behavior of a real API and language models (LLMs) can interpret the contract without ambiguity.

The main goals are:

1. **Standardization and versioning** – A `.apimock` file describes an HTTP route using plain text. It can be versioned alongside your source code and serves as the source of truth for the contract.
2. **Mock API support** – A mock server reads the file, validates request data based on the schema defined in the request section, and selects the appropriate response based on status codes and properties.
3. **LLM‑friendly** – The format is simple yet semantically rich; language models can infer API behavior by analyzing the structured sections and body schemas.

Each `.apimock` file contains **two parts**: the **request definition** and one or more **responses**. These parts are separated by a blank line. The syntax and usage guidelines are presented below.

## Request definition

The request section describes how the client must invoke the route. It consists of the following lines, in order:

```
METHOD path
Optional properties…

optional‑body‑schema
```

### Method and path (`METHOD path`)

* **Method (optional)** – Must be a valid HTTP method (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, etc.). The HTTP specification defines what each method means and influences the semantics of the response (for example, `201 Created` is typically returned after a successful `POST`).
* **Path (required)** – Defines the resource route. It uses **placeholders** in curly braces `{}` for path parameters, similar to OpenAPI: variable segments are written as `{name}` and must be replaced by an actual value when making the call. A path can contain multiple parameters, for example `/cars/{carId}/drivers/{driverId}`.

  * Static segments and placeholders can be split across **multiple lines** for readability. The first line must contain the method and the beginning of the path; subsequent lines that start with `/` or `?` are concatenated in the same order. This lets you clearly separate segments and query parameters, as in the example:

    ```apimock
    GET /api
        /v1
        /users
        /{userId}
        /posts
        ?page=1
        &limit=10
        &sort=created_at
    ```

### Request properties

After the `METHOD path` line, you can declare key:value properties related to HTTP request headers. Each property occupies its own line and is separated by a colon. Two properties are most common:

* **`Accept:`** – Tells the server which media types (MIME types) are accepted in the **response body**. According to MDN, the `Accept` header lists the MIME types the agent can process and may include quality factors for prioritization.  If unspecified, any response type is considered acceptable. Examples: `Accept: application/json`, `Accept: application/xml`.

* **`ContentType:`** (equivalent to the HTTP `Content-Type` header) – Indicates the media type of the **request body**. The `Content-Type` header specifies the format of the content sent in `POST`/`PUT`/`PATCH` requests. Examples: `ContentType: application/json`, `ContentType: text/plain;charset=utf-8`, `ContentType: multipart/form-data`.

Other header properties can be supported by extensions, but the basic specification focuses on these two headers. Unknown properties should be ignored by the mock server.

### Body schema (optional)

If the route expects a body, you can define a **validation schema** immediately after the request properties. This schema acts as a contract for the data being sent and allows the mock server to return an appropriate error (e.g. a `400 Bad Request`) when the body does not conform to the schema. The schema format depends on the `ContentType` value:

1. **JSON Schema** – For JSON, plain text, or YAML content, use the [JSON Schema](https://json-schema.org) vocabulary. JSON Schema is a standardized vocabulary used to annotate and validate JSON documents, providing metadata and constraints on types and properties. A JSON schema is a JSON object containing keys like `type`, `properties`, `required`, etc.

   Example schema for a user JSON body:

   ```json
   {
     "title": "User",
     "type": "object",
     "properties": {
       "name": { "type": "string" },
       "age":  { "type": "number" }
     },
     "required": ["name", "age"]
   }
   ```

2. **XML Schema (XSD)** – For XML content, use the XSD standard. XSD is a W3C recommendation for describing and validating the structure and content of XML documents; it defines allowed elements, attributes and data types. An XSD schema is written in XML and can be included directly after the request definition.

   Simplified example:

   ```xml
   <xs:schema>
     <xs:element name="user" type="xs:string"/>
   </xs:schema>
   ```

3. **Other formats** – For specialized media types (e.g. `multipart/form-data`, `application/x-www-form-urlencoded`, `application/octet-stream`), the JSON schema may include additional keywords (`format: "binary"`, `contentMediaType`, etc.) to describe file uploads.

If no schema is defined, the body is considered free‑form and the mock server will not perform validation. Omitting the schema also means the route can accept any HTTP method or path if `METHOD` and `path` are empty, making the file generic.

## Response definition

After the request section, define one or more **responses**. Each response begins with a line that starts with `--` and follows this structure:

```
-- status-code: optional description
Optional properties…

response-body
```

### Opening line

The `--` marker signals the start of a new response. It is followed by a **status code** (required) and optionally a brief description. Status codes follow the HTTP specification; for example, `200` indicates success, `201` means a resource was created and `4xx` indicates client errors. You can define multiple responses with the same status code to test different scenarios (for example, different messages for `400 Bad Request`).

### Response properties

Next, you can list key:value properties that customise the mock server's behaviour:

* **`ContentType:`** – Defines the MIME type of the response body. According to MDN, the `Content-Type` header informs the client about the media type of the returned data. If not defined, it defaults to `text/plain`. The value may include parameters such as `charset=utf-8`.

* **`Delay:`** – An integer representing the number of milliseconds the mock server should wait before sending the response. Useful for simulating network latency or processing time.

Unknown properties should be ignored, allowing custom extensions in future implementations.

### Response body

After a blank line, include the response body. The content must match the type defined in `ContentType`. Common types include:

* `application/json` – The body must be valid JSON.
* `application/xml` or `text/xml` – The body must be valid XML.
* `text/plain` – Any plain text.
* `text/yaml` – YAML content.
* For binary data (`application/octet-stream`), you may include a base64‑encoded value or a textual description agreed upon between the mock and the consumer.

When multiple responses are defined, the choice of which response to return may depend on the mock server's logic (e.g. return the first matching response, choose based on a header), or may be randomised for testing.

### Full example

The file below demonstrates the combination of several concepts:

```apimock
POST /user/{userId}/image
Accept: multipart/form-data
ContentType: multipart/form-data
{
  "title": "User Image Upload",
  "type": "object",
  "properties": {
    "image": {
      "type": "string",
      "format": "binary",
      "description": "Image file (JPEG, PNG, GIF, WebP)"
    },
    "alt_text": {
      "type": "string",
      "description": "Alternate text for the image",
      "maxLength": 255
    }
  },
  "required": ["image"]
}

-- 200: Image updated
ContentType: application/json
Delay: 200

{
  "message": "Image updated successfully",
  "image_url": "https://example.com/images/123.jpg"
}

-- 415: Invalid file type
ContentType: application/json

{
  "message": "Unsupported media type. Only JPEG, PNG, GIF, and WebP are allowed"
}
```

## Conventions and considerations

### MIME types

A MIME type (also called a media type) indicates the nature and format of a document and is composed of a **type** and a **subtype** separated by `/`. The Internet Assigned Numbers Authority (IANA) maintains the official registry of MIME types. It is important that the `Content-Type` sent in responses is consistent, because browsers use this value to process the content correctly. Examples:

| MIME type                  | Usage                                                |
| -------------------------- | ---------------------------------------------------- |
| `application/json`         | Structured JSON data for APIs                        |
| `application/xml`          | Structured XML data                                  |
| `text/plain`               | Plain text; use `charset` to indicate encoding       |
| `text/yaml`                | YAML documents                                       |
| `application/octet-stream` | Arbitrary binary data                                |
| `multipart/form-data`      | Used for multipart uploads such as form file uploads |

### Relevant HTTP headers

The `Accept` header indicates the media types the agent can process and may include a quality factor (q‑value) for prioritisation. The `Content-Type` header indicates the media of the request or response body. Both should be used according to HTTP content negotiation to ensure interoperability and avoid errors such as `415 Unsupported Media Type`.

### Path and query parameters

* Path parameters are indicated by curly braces `{}` in the path and are **required**. Each placeholder must have a corresponding value in the actual call.
* Query parameters are appended after the question mark `?` and separated by `&`, for example `?page=1&limit=10`. They may be optional depending on the API.

### Validation and errors

If a request body does not conform to the defined schema, the mock server should return an appropriate response (for example, `400 Bad Request` or `422 Unprocessable Entity`) rather than the success responses. Error messages can be defined as additional responses in the file.

### Comments and whitespace

Blank lines can be used to separate sections and improve readability; mock servers and LLMs should ignore them when interpreting the document. At present the **apimock** language does not define an official comment syntax; any notes should be written in the description of responses or in a separate document.

## Guidelines for tools and language models

To make the format easier for humans and LLMs:

1. **Use consistent names** – Properties (e.g. `Accept`, `ContentType`, `Delay`) should respect the capitalisation shown; LLMs may differentiate if there is variation.
2. **Provide complete schemas** – Whenever possible, supply a detailed schema (JSON Schema or XSD). This helps tools validate inputs and aids language models in generating coherent examples.
3. **Avoid long text in tables** – Use tables only for keywords, numbers or short descriptions. Long prose should remain in the main text.
4. **Respect the order of sections** – The request comes first, followed by a blank line and then the responses. LLMs can use this order to segment the document.
5. **Treat path variables as literals** – When reading an `.apimock` file, treat placeholders inside `{}` as variables and do not attempt to interpret them until concrete values are available. This avoids confusing LLMs generating example requests.

## References

* **HTTP status codes** – MDN notes that status codes indicate whether a request has been successfully completed and groups them into five classes (100–199 informational, 200–299 success, 300–399 redirection, 400–499 client errors, 500–599 server errors).
* **JSON Schema** – The JSON Schema specification defines a vocabulary for annotating and validating JSON documents.
* **MIME types** – A media type indicates the format of a document and is composed of `type/subtype`; the IANA maintains the official registry. Browsers rely on the `Content-Type` header to interpret responses correctly.
* **Accept header** – The `Accept` header lists the MIME types the agent accepts and may include quality factors to indicate preference.
* **Path placeholders** – In APIs, path parameters are variable parts of a URL, delimited by curly braces `{}`.
* **XML Schema Definition (XSD)** – The W3C recommends XSD for describing and validating the structure of XML documents.
