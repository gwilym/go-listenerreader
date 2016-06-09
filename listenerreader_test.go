package listenerreader

import (
	"bufio"
	"net"
	"sort"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type ListenerReaderSuite struct{}

var _ = Suite(&ListenerReaderSuite{})

func (s *ListenerReaderSuite) TestSomething(c *C) {
	addr := "localhost:3333"

	listener, err := net.Listen("tcp", addr)
	c.Assert(err, IsNil)
	defer listener.Close()

	lr := NewListenerReader(listener, '\n', 1, 64*1024, 1)
	defer lr.Close()

	datach := make(chan string)

	for i := 0; i < 3; i++ {
		go func() {
			conn, err := net.Dial("tcp", addr)
			c.Assert(err, IsNil)
			defer conn.Close()
			var prev bool
			for line := range datach {
				if prev {
					n, err := conn.Write([]byte{'\n'})
					c.Assert(n, Equals, 1)
					c.Assert(err, IsNil)
				} else {
					prev = true
				}
				n, err := conn.Write([]byte(line))
				c.Assert(n, Equals, len(line))
				c.Assert(err, IsNil)
			}
		}()
	}

	data := []string{
		/* 756 */ "Cras dictum nisi nec urna scelerisque mattis. Aliquam vitae risus quis diam pulvinar congue a sed arcu. Curabitur aliquet malesuada lorem et fermentum. Praesent ipsum urna, feugiat ac blandit quis, luctus eget massa. Duis sollicitudin turpis augue, sit amet feugiat nisi dictum vitae. Curabitur molestie ut lectus et ornare. Sed blandit, purus non semper congue, nunc arcu lobortis elit, vel rhoncus urna ligula vitae lorem. Phasellus nec scelerisque nisl. Integer mi leo, pellentesque non interdum ac, vehicula nec nunc. Mauris egestas neque semper augue interdum egestas. Proin viverra lacus sit amet enim sagittis, consectetur pharetra erat vehicula. Morbi malesuada quam eget libero hendrerit, sed hendrerit nisl semper. Pellentesque et fringilla odio.",
		/* 330 */ "Donec nec consectetur lacus. Aliquam luctus laoreet orci sit amet malesuada. Etiam nisi velit, varius sit amet viverra a, posuere et libero. Duis quis dolor neque. In elementum lobortis tristique. Cras nulla ante, fermentum ac massa ac, fringilla bibendum arcu. Nam eleifend magna in orci tempor pulvinar. Aenean nec libero velit.",
		/* 627 */ "Duis mauris elit, luctus vel justo sed, mattis faucibus dolor. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Suspendisse euismod neque a mollis suscipit. Donec neque lacus, consequat vitae lectus a, consequat dapibus orci. Sed interdum libero nisl, aliquam laoreet sem dapibus ac. Nullam viverra augue eu mi vehicula pharetra. Donec arcu felis, ornare nec condimentum vel, consectetur id purus. Praesent maximus elementum est, nec auctor libero mattis in. Mauris at ultricies tortor. Sed vel lobortis ipsum. Mauris rutrum porttitor lectus eu accumsan. In hac habitasse platea dictumst.",
		/* 640 */ "Etiam sed erat dictum, tempor eros vitae, ornare arcu. Donec quis dui et enim aliquam eleifend sit amet a justo. Morbi feugiat mi id enim interdum mollis. Mauris auctor purus ut nisl sagittis, eu molestie turpis pellentesque. Cras et placerat nulla. Mauris faucibus magna quis sem dignissim, sit amet bibendum urna posuere. Sed et rhoncus dolor. Integer dignissim, dui nec lacinia feugiat, mi elit iaculis justo, aliquet iaculis nisl metus a turpis. In vulputate pharetra mi quis molestie. Donec vitae nibh purus. In hac habitasse platea dictumst. In a urna a dolor accumsan volutpat id eu ligula. Proin eget metus ut est sodales imperdiet.",
		/* 674 */ "Fusce metus leo, blandit et nisl sed, vestibulum tincidunt velit. Etiam a fringilla ante. Nullam lacinia nisi sed gravida semper. Fusce mollis ex quis metus volutpat lacinia. Aliquam egestas fringilla pretium. Vivamus imperdiet consectetur magna, sed pellentesque leo pharetra in. In non purus in massa efficitur auctor. Sed dapibus lacinia massa, at dignissim sapien tristique sit amet. Nam eu urna rutrum, tincidunt nulla et, dictum nulla. Suspendisse tincidunt nibh a maximus volutpat. Donec imperdiet, neque convallis congue congue, dui massa convallis metus, id fringilla tortor lectus gravida odio. Aenean ut purus convallis, venenatis est sit amet, pellentesque nunc.",
		/* 462 */ "Fusce tristique vehicula leo ac volutpat. Curabitur vitae nisl lectus. Sed in tortor vel ipsum venenatis tincidunt et eget nibh. Curabitur in convallis sapien. Ut lacus tortor, tristique hendrerit ultricies ac, tincidunt sit amet eros. Curabitur tincidunt accumsan arcu, a laoreet sem luctus id. Morbi interdum enim vel elementum egestas. Suspendisse ex est, euismod ac ex a, viverra aliquam metus. Cras dolor felis, sodales ut purus non, elementum porta libero.",
		/* 672 */ "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec elementum neque dui, et posuere tortor sollicitudin at. Praesent pretium odio non varius posuere. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Integer quis tincidunt tellus, ac ultricies purus. Duis eget turpis nec velit porttitor accumsan. Proin aliquet dapibus risus. Cras tristique elementum mauris sit amet laoreet. Duis egestas, eros ultrices lacinia pretium, risus mauris laoreet mauris, et dapibus dui risus quis justo. Sed sit amet arcu at lacus posuere efficitur eget quis ante. Nulla malesuada lacus eget massa gravida, finibus feugiat nisl vestibulum.",
		/* 847 */ "Nulla interdum, mi sit amet efficitur consectetur, ipsum neque pellentesque nulla, sit amet consequat ipsum velit non neque. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Cras facilisis condimentum justo. Etiam in lorem vel lacus semper porttitor. Duis dignissim neque in fermentum efficitur. Proin gravida a diam in accumsan. Suspendisse magna urna, fermentum eget nisi gravida, ornare tristique felis. Quisque gravida tellus ligula, non viverra justo ullamcorper convallis. Proin eget facilisis libero, a tincidunt lorem. Integer sagittis, lectus non tincidunt maximus, ex est efficitur tellus, ac venenatis sapien lacus eu est. Aenean mollis arcu lorem, non rutrum sem consectetur quis. Aliquam erat volutpat. Phasellus sit amet ipsum metus.",
		/* 874 */ "Quisque finibus luctus ullamcorper. Suspendisse sapien felis, ultricies malesuada ultricies commodo, sodales et mauris. Etiam sed lorem nec justo placerat blandit. Vivamus vulputate fringilla sem. Aenean nibh erat, varius ut eros vitae, congue accumsan dolor. Donec id faucibus felis. Nunc dapibus massa imperdiet, lacinia nunc sit amet, semper nunc. Mauris faucibus non sapien vitae dictum. Etiam lacus sem, dictum at lacinia in, sodales ultrices augue. In vehicula semper turpis. Aliquam posuere eros eget elit finibus, eget sodales mauris fermentum. Cras euismod lectus felis, ac aliquam lacus sollicitudin a. In laoreet tempus massa, nec tempus libero finibus mollis. In leo mi, vulputate sit amet eleifend et, pharetra at odio. Nullam rhoncus tincidunt condimentum. Vestibulum maximus, ante ac tristique gravida, nunc ex tempus lacus, quis vulputate eros ex nec lectus.",
	}
	sort.Strings(data)

	for _, line := range data {
		datach <- line
	}
	close(datach)

	scanner := bufio.NewScanner(lr)
	scanned := []string{}

Loop:
	for {
		ch := make(chan bool)
		go func() {
			defer close(ch)
			ch <- scanner.Scan()
		}()

		select {
		case <-ch:
			scanned = append(scanned, string(scanner.Bytes()))
			c.Assert(scanner.Err(), IsNil)
			if len(scanned) >= len(data) {
				break Loop
			}
		case <-time.After(100 * time.Millisecond):
			break Loop
		}
	}

	sort.Strings(scanned)
	c.Assert(scanned, DeepEquals, data)
}
