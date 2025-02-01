# jcllm: A Command-Line Interface for Large Language Models

![jcllm screencast](https://binaries.jchengdev.org/i/jcllm-0002.gif)


For developers who live in the terminal, `jcllm` offers a convenient and efficient way to access large language models. While web interfaces like ChatGPT are powerful, `jcllm` provides a streamlined alternative for quick queries and tasks without leaving the command line. This tool prioritizes responsiveness, delivering results in seconds when configured to use fast models.

# Installation

Go to https://github.com/jlcheng/jcllm/releases and download the archive for your operating system. The archive will contain a single executable named `jcllm`.

# Configuration

**1. You can run jcllm without any configuration, by specifying required options on the command-line.**

```
jcllm --provider gemini --model=gemini-1.5-flash-8b --gemini-api-key=...
```

**2. You can also set options as environment variables, to avoid specifying the API key on the command-line.**

```
export JCLLM_GEMINI_API_KEY=...

jcllm --provider gemini --model=gemini-1.5-flash-8b
```

> Any CLI argument can be turned into an environmnet variable by 1) Adding "JCLLM_" as a prefix; 2) Making the string uppercase; and 3) Changing dashes into underscores. 
> e.g., `--provider` becomes `JCLLM_PROVIDER` and `--gemini-api-key` becomes `JCLLM_GEMINI_API_KEY`.

**3. You can also use a configuration file.**

The program will look for a configuration file in the local directory named `.jcllm.toml`, or `$HOME/.jcllm.d/jcllm.toml`.

```
provider="gemini"
model="gemini-1.5-flash-8b"

gemini-api-key="..."
gemini-models-list=[
  "gemini-1.5",
  "gemini-1.5-flash-8b",
  "gemini-2.0-flash-exp",
  "gemini-2.0-flash-thinking-exp", 
]

openai-base-url=https://api.openai.com/v1
openai-api-key="..."
openai-models-list=[
  "gpt-4o",
  "o1-mini",
  "o1-preview",
]


system-prompt="""
Talk like a pirate.
"""
```
