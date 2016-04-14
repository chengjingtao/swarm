package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
    "fmt"
)

func TestCheckAddrFormat(t *testing.T) {
    //t.Fail()
	assert.False(t, checkAddrFormat("1.1.1.1"))
	assert.False(t, checkAddrFormat("hostname"))
	assert.False(t, checkAddrFormat("1.1.1.1:"))
	assert.False(t, checkAddrFormat("hostname:"))
	assert.False(t, checkAddrFormat("1.1.1.1:111111"))
	assert.False(t, checkAddrFormat("hostname:111111"))
	assert.False(t, checkAddrFormat("http://1.1.1.1"))
	assert.False(t, checkAddrFormat("http://hostname"))
	assert.False(t, checkAddrFormat("http://1.1.1.1:1"))
	assert.False(t, checkAddrFormat("http://hostname:1"))
	assert.False(t, checkAddrFormat(":1.1.1.1"))
	assert.False(t, checkAddrFormat(":hostname"))
	assert.False(t, checkAddrFormat(":1.1.1.1:1"))
	assert.False(t, checkAddrFormat(":hostname:1"))
	assert.True(t, checkAddrFormat("1.1.1.1:1111"))
	assert.True(t, checkAddrFormat("hostname:1111"))
	assert.True(t, checkAddrFormat("host-name_42:1111"))
}

func TestRefreshURL(t *testing.T){
    var url="192.168.5.55:2375"
    u := refreshURL(url)
    fmt.Println(u)
    if u !="192.168.5.55:2375"{
        t.Error("not 192.168.5.55:2375")
    }
    
    
    u =  refreshURL("192.168.5.55/weight")
    fmt.Println(u)
    
    if u !="192.168.5.55/weight"{
        t.Error("not 192.168.5.55/weight")
    }
    
    u =  refreshURL("192.168.5.55/weight/0")
    fmt.Println(u)
    
    if u !="192.168.5.55/weight/0"{
        t.Error("not 192.168.5.55/weight/0")
    }
    
    u =  refreshURL("192.168.5.55/weight/1")
    t.Log(u)
    
    if u !="192.168.5.55/weight/1"{
        t.Error("not 192.168.5.55/weight/1")
    }
}

func TestRefreshURLWithENVWeight(t *testing.T){
    var envW="-1"
    
    var url="192.168.5.55:2375"
    u := refreshURL(url)
    fmt.Println(u)
    if u !="192.168.5.55:2375/weight/"+envW{
        t.Error("not 192.168.5.55:2375/weight/"+envW)
    }
    
    
    u =  refreshURL("192.168.5.55/weight")
    fmt.Println(u)
    if u !="192.168.5.55/weight/"+envW{
        t.Error("not 192.168.5.55/weight/"+envW)
    }
    
    u =  refreshURL("192.168.5.55/weight/0")
    fmt.Println(u)
    if u !="192.168.5.55/weight/"+envW{
        t.Error("not 192.168.5.55/weight/"+envW)
    }
    
    u =  refreshURL("192.168.5.55/weight/9")
    if u !="192.168.5.55/weight/"+envW{
        t.Error("not 192.168.5.55/weight/"+envW)
    }
}
