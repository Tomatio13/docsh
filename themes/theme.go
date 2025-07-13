package themes

import (
	"fmt"
	"strings"
)

type Theme struct {
	Name        string
	Prompt      string
	Colors      map[string]string
	Symbols     map[string]string
}

var BuiltinThemes = map[string]*Theme{
	"default": {
		Name:   "default",
		Prompt: "cherry:%s$ ",
		Colors: map[string]string{
			"reset":   "\033[0m",
			"red":     "\033[31m",
			"green":   "\033[32m",
			"yellow":  "\033[33m",
			"blue":    "\033[34m",
			"magenta": "\033[35m",
			"cyan":    "\033[36m",
			"white":   "\033[37m",
			"bold":    "\033[1m",
		},
		Symbols: map[string]string{
			"arrow":  "➜",
			"branch": "⎇",
			"dirty":  "✗",
			"clean":  "✓",
		},
	},
	"robbyrussell": {
		Name:   "robbyrussell",
		Prompt: "%{$fg[cyan]%}%s %{$fg[red]%}➜ %{$reset_color%}",
		Colors: map[string]string{
			"reset":   "\033[0m",
			"red":     "\033[31m",
			"green":   "\033[32m",
			"yellow":  "\033[33m",
			"blue":    "\033[34m",
			"magenta": "\033[35m",
			"cyan":    "\033[36m",
			"white":   "\033[37m",
			"bold":    "\033[1m",
		},
		Symbols: map[string]string{
			"arrow":  "➜",
			"branch": "",
			"dirty":  "✗",
			"clean":  "",
		},
	},
	"agnoster": {
		Name:   "agnoster",
		Prompt: "%{$fg[green]%}%u@%h%{$reset_color%} %{$fg[blue]%}%s%{$reset_color%} $ ",
		Colors: map[string]string{
			"reset":   "\033[0m",
			"red":     "\033[31m",
			"green":   "\033[32m",
			"yellow":  "\033[33m",
			"blue":    "\033[34m",
			"magenta": "\033[35m",
			"cyan":    "\033[36m",
			"white":   "\033[37m",
			"bold":    "\033[1m",
		},
		Symbols: map[string]string{
			"arrow":  "⮀",
			"branch": "⎇",
			"dirty":  "±",
			"clean":  "",
		},
	},
	"simple": {
		Name:   "simple",
		Prompt: "%s $ ",
		Colors: map[string]string{
			"reset": "\033[0m",
		},
		Symbols: map[string]string{},
	},
}

func GetTheme(name string) (*Theme, bool) {
	theme, exists := BuiltinThemes[name]
	return theme, exists
}

func (t *Theme) ApplyColors(text string) string {
	result := text
	
	// oh-my-zsh形式のカラープレースホルダーを置換
	for colorName, colorCode := range t.Colors {
		placeholder := fmt.Sprintf("$fg[%s]", colorName)
		result = strings.ReplaceAll(result, placeholder, colorCode)
		
		placeholder = fmt.Sprintf("%%{$fg[%s]%%}", colorName)
		result = strings.ReplaceAll(result, placeholder, colorCode)
	}
	
	// リセットカラー
	result = strings.ReplaceAll(result, "$reset_color", t.Colors["reset"])
	result = strings.ReplaceAll(result, "%{$reset_color%}", t.Colors["reset"])
	result = strings.ReplaceAll(result, "%{reset%}", t.Colors["reset"])
	
	// 未解決のカラープレースホルダーを除去（空白問題の解決）
	result = t.removeUnresolvedPlaceholders(result)
	
	return result
}

func (t *Theme) removeUnresolvedPlaceholders(text string) string {
	result := text
	
	// 未解決の %{...%} 形式のプレースホルダーを除去
	for {
		start := strings.Index(result, "%{")
		if start == -1 {
			break
		}
		
		end := strings.Index(result[start:], "%}")
		if end == -1 {
			break
		}
		
		end = start + end + 2 // "%}" の分を加算
		result = result[:start] + result[end:]
	}
	
	// 未解決の $fg[...] 形式のプレースホルダーを除去
	for {
		start := strings.Index(result, "$fg[")
		if start == -1 {
			break
		}
		
		end := strings.Index(result[start:], "]")
		if end == -1 {
			break
		}
		
		end = start + end + 1 // "]" の分を加算
		result = result[:start] + result[end:]
	}
	
	// $reset_color の未解決分を除去
	result = strings.ReplaceAll(result, "$reset_color", "")
	
	return result
}

func (t *Theme) GetPrompt() string {
	return t.ApplyColors(t.Prompt)
}

func ListThemes() {
	fmt.Println("Available themes:")
	for name, theme := range BuiltinThemes {
		fmt.Printf("  %s - %s\n", name, theme.Name)
	}
}