# jcllm: A Command-Line Interface for Large Language Models

For developers who live in the terminal, `jcllm` offers a convenient and efficient way to access large language models. While web interfaces like ChatGPT are powerful, `jcllm` provides a streamlined alternative for quick queries and tasks without leaving the command line. This tool prioritizes responsiveness, delivering results in seconds when configured to use fast models.

# Installation

Go to https://github.com/jlcheng/jcllm/releases and download the archive for your operating system. The archive will contain a single executable named `jcllm`.

# Configuration

1. You can run jcllm without any configuration

```
jcllm --provider gemini --model=gemini-1.5-flash-8b --gemini-api-key=...
```

2. You can specify api keys via environment variables

```
export JCLLM_GEMINI_API_KEY=...

jcllm --provider gemini --model=gemini-1.5-flash-8b
```

Any option on the command line can be converted into an environmnet variable by adding "JCLLM_", making the option uppercase, and changing the dash into an underscore. 

e.g., proviedr => `JCLLM_PROVIDER`, gemini-api-key => 'JCLLM_GEMINI_API_KEY`.

3. You can use a configuratoin file.

The program will look for a configuration file in the local directory named `.jcllm.toml`, or `$HOME/.jcllm.d/jcllm.toml`.

```
provider="gemini"
model="gemini-1.5-flash-8b"
gemini-api-key="..."

system-prompt="""
Talk like a pirate.
"""
```
