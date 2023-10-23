# Golang GODOT decoder

A basic decoder in golang for GODOT decoder.

_I create this repository as draft and i do not maintain it so fork or copy some
part._

## File format

Source: https://github.com/Bioruebe/godotdec#technical-details

| Value/Type | Description                                        |
| ---------: | -------------------------------------------------- |
|            | --- Package Information                            |
| 0x43504447 | Magic number "GDPC"                                |
|  4 * int32 | Engine: version, major, minor, revision            |
| 16 * int32 | Reserved space, 0                                  |
|      int32 | Number of files in archive                         |
|            | --- Array of file informations                     |
|      int32 | Length of path string                              |
|     string | Path as string, e.g. res://actors/Enemy/enemy.atex |
|      int64 | File offset from file begin                        |
|      int64 | File size                                          |
|   16*bytes | MD5                                                |
|            | --- All the files content                          |
