package gotcp

import "syscall"

func EpollAddFd(epollfd, fd int, flag int) error {
	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: uint32(flag),
	}
	return syscall.EpollCtl(epollfd, syscall.EPOLL_CTL_ADD, fd, event)
}

func EpollModFd(epollfd, fd int, flag int) error {
	event := &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: uint32(flag),
	}
	return syscall.EpollCtl(epollfd, syscall.EPOLL_CTL_MOD, fd, event)
}

func EpollDelFd(epollfd, fd int) error {
	event := &syscall.EpollEvent{
		Fd: int32(fd),
	}
	return syscall.EpollCtl(epollfd, syscall.EPOLL_CTL_DEL, fd, event)
}
