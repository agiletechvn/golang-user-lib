package stringutil

func StringArgsToBytesArgs(args []string) [][]byte {
  a := make([][]byte, len(args))
  for k, v := range args {
    a[k] = []byte(v)
  }
  return a
}
