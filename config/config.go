package config

import (
    //"fmt"
    //"reflect"

    "github.com/spf13/viper"
    "github.com/gdamore/tcell/v2"
    flag "github.com/spf13/pflag"
    "code.rocketnine.space/tslocum/cbind"
)

func Init(flagList *flag.FlagSet) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.SetDefault("db-file", "~/.geek-life/default.db")
    viper.SetDefault("vertical", false)
    viper.SetDefault("dynamic", false)
    viper.SetDefault("upKeys", []string{"j", "Up"})
    viper.SetDefault("downKeys", []string{"k", "Down"})
    //viper.AddConfigPath("$HOME/.config/geek-life")
    viper.AddConfigPath("$HOME/.geek-life")
    viper.BindPFlags(flagList)
    viper.ReadInConfig()
}

func SaveConfig() {
    viper.SafeWriteConfig()
    //if _, ok := err.(viper.ConfigFileNotFoundError); ok {
        //viper.SafeWriteConfigAs("$HOME
    //}
}

func GetDbFile() string {
    return viper.GetString("db-file")
}

func GetVertical() bool {
    return viper.GetBool("vertical")
}

func GetDynamic() bool {
    return viper.GetBool("dynamic")
}

func GetUpKeysAsString() []string {
    return viper.GetStringSlice("upKeys")
}

func GetDownKeysAsString() []string {
    return viper.GetStringSlice("downKeys")
}

func GetRunes(allKeys []string) ([]rune, []tcell.Key) {
    upRunes := make([]rune, 0, len(allKeys))
    upKeys := make([]tcell.Key, 0, len(allKeys))
    for i := 0; i < len(allKeys); i++{
        allRunes := []rune(allKeys[i])
        if 1 == len(allRunes) {
            upRunes = append(upRunes, allRunes[0])
        } else {
            _, key, _, _ := cbind.Decode(allKeys[i])
            upKeys = append(upKeys, key)
        }
    }
    
    return upRunes, upKeys
}

func GetUpRunes() []rune{
    runes, _ := GetRunes(GetUpKeysAsString())
    return runes
}

func GetUpKeys() []tcell.Key{
    _, keys := GetRunes(GetUpKeysAsString())
    return keys
}

func GetDownRunes() []rune{
    runes, _ := GetRunes(GetDownKeysAsString())
    return runes
}

func GetDownKeys() []tcell.Key{
    _, keys := GetRunes(GetDownKeysAsString())
    return keys
}
