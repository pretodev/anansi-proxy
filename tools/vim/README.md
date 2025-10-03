# ApiMock Syntax Highlighting for Vim/Neovim

Syntax highlighting plugin for `.apimock` files used by Anansi Proxy.

## Installation

### Using vim-plug

Add to your `.vimrc` or `init.vim`:

```vim
Plug '~/path/to/anansi-proxy/tools/vim'
```

Or if you have the repo cloned:

```vim
Plug 'pretodev/anansi-proxy', {'rtp': 'tools/vim'}
```

Then run:
```
:PlugInstall
```

### Using packer.nvim (Neovim)

Add to your `init.lua`:

```lua
use {
  'pretodev/anansi-proxy',
  rtp = 'tools/vim'
}
```

### Manual Installation

Copy the files to your vim/neovim runtime path:

```bash
# For Vim
cp -r syntax ftdetect ~/.vim/

# For Neovim
cp -r syntax ftdetect ~/.config/nvim/
```

## Features

- Syntax highlighting for HTTP methods
- Status code highlighting (200, 404, etc.)
- Response separator highlighting (`--`)
- Path parameter highlighting (`{param}`)
- Header/property highlighting
- Comment support with `#`
- JSON body highlighting
- XML body highlighting
- YAML body highlighting

## Usage

Files with `.apimock` extension will automatically be detected and highlighted.

You can also manually set the filetype:
```vim
:set filetype=apimock
```

## License

MIT
