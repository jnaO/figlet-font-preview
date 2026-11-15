# bash completion for figlet-font-preview (standalone; works on bash 3.2)
_figlet_font_preview_reply() {
    local line
    COMPREPLY=()
    while IFS= read -r line; do
        COMPREPLY+=("$line")
    done < <(compgen "$@")
}

_figlet_font_preview() {
    local cur=${COMP_WORDS[COMP_CWORD]} prev=
    if [ "$COMP_CWORD" -gt 0 ]; then
        prev=${COMP_WORDS[COMP_CWORD-1]}
    fi

    # With "=" in COMP_WORDBREAKS, "--font-dir=DIR" arrives as three words.
    if [ "$cur" = "=" ] && [ "$COMP_CWORD" -gt 0 ]; then
        cur=
        prev=${COMP_WORDS[COMP_CWORD-1]}
    elif [ "$prev" = "=" ] && [ "$COMP_CWORD" -gt 1 ]; then
        prev=${COMP_WORDS[COMP_CWORD-2]}
    fi

    case $prev in
        --font-dir | -font-dir)
            _figlet_font_preview_reply -d -- "$cur"
            return 0
            ;;
    esac

    case $cur in
        --font-dir=*)
            _figlet_font_preview_reply -d -- "${cur#--font-dir=}"
            return 0
            ;;
        -*)
            _figlet_font_preview_reply -W '--center --right --font-dir --help --version' -- "$cur"
            ;;
        *)
            COMPREPLY=()
            ;;
    esac
    return 0
}

complete -o filenames -F _figlet_font_preview figlet-font-preview
