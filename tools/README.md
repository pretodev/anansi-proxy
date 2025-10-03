# ApiMock Editor Support

This directory contains syntax highlighting plugins for various code editors.

## Available Editors

### VS Code
See [vscode/README.md](vscode/README.md) for installation instructions.

**Quick Install:**
```bash
cd tools/vscode
code --install-extension .
```

### Vim/Neovim
See [vim/README.md](vim/README.md) for installation instructions.

**Quick Install (Manual):**
```bash
# For Vim
cp -r tools/vim/syntax tools/vim/ftdetect ~/.vim/

# For Neovim
cp -r tools/vim/syntax tools/vim/ftdetect ~/.config/nvim/
```

## Features

Both plugins provide:
- HTTP method highlighting (GET, POST, PUT, DELETE, etc.)
- Status code highlighting (200, 404, 500, etc.)
- Response separator highlighting (`--`)
- Path parameter highlighting (`{userId}`, `{id}`, etc.)
- Property/header highlighting
- Comment support (`#`)
- String and number highlighting
- Embedded JSON/XML/YAML support

## Contributing

Feel free to submit issues or pull requests to improve the syntax highlighting for any editor.

## Other Editors

Contributions for other editors are welcome:
- Sublime Text
- IntelliJ IDEA
- Emacs
- Atom
- etc.
