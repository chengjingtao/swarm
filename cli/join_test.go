package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
    _ "fmt"
)

func TestCheckAddrFormat(t *testing.T) {
    //t.Fail()
	// assert.False(t, checkAddrFormat("1.1.1.1"))
	// assert.False(t, checkAddrFormat("hostname"))
	// assert.False(t, checkAddrFormat("1.1.1.1:"))
	// assert.False(t, checkAddrFormat("hostname:"))
	// assert.False(t, checkAddrFormat("1.1.1.1:111111"))
	// assert.False(t, checkAddrFormat("hostname:111111"))
	// assert.False(t, checkAddrFormat("http://1.1.1.1"))
	// assert.False(t, checkAddrFormat("http://hostname"))
	// assert.False(t, checkAddrFormat("http://1.1.1.1:1"))
	// assert.False(t, checkAddrFormat("http://hostname:1"))
	// assert.False(t, checkAddrFormat(":1.1.1.1"))
	// assert.False(t, checkAddrFormat(":hostname"))
	// assert.False(t, checkAddrFormat(":1.1.1.1:1"))
	// assert.False(t, checkAddrFormat(":hostname:1"))
	// assert.True(t, checkAddrFormat("1.1.1.1:1111"))
	// assert.True(t, checkAddrFormat("hostname:1111"))
	// assert.True(t, checkAddrFormat("host-name_42:1111"))
}

func TestRefreshData(t *testing.T){
    var url="192.168.5.55:2375"
    u := refreshData(url)
   assert.True(t,len(u)==0)
    
    
    u =  refreshData("192.168.5.55?weight")
    assert.True(t,u["weight"]=="")
    
    u =  refreshData("192.168.5.55?weight=0")
    assert.True(t,u["weight"]=="0")
    
    u =  refreshData("192.168.5.55?weight=1")
    assert.True(t,u["weight"]=="1")
}

func TestRefreshDataWithENVWeight(t *testing.T){
    var envW="-1"
    
    var url="192.168.5.55:2375"
    u := refreshData(url)
    assert.True(t,u["weight"]==envW)
    
    
    u =  refreshData("192.168.5.55?weight")
    assert.True(t,u["weight"]==envW)
    
    u =  refreshData("192.168.5.55?weight=0")
    assert.True(t,u["weight"]==envW)

    
    u =  refreshData("192.168.5.55?weight=9")
    assert.True(t,u["weight"]==envW)
}
