package main

func getMapKeys[K comparable,V any](mymap map[K]V) []K {
    keys := make([]K, len(mymap))

    i := 0
    for k := range mymap {
        keys[i] = k
        i++
    }

    return keys
}

