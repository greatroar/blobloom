// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !appengine,gc,!purego

#include "textflag.h"

// SSE2-enabled union of two slices of blocks, assumed to be of the same length.
// func union(a, b []block)
TEXT ·union(SB), NOSPLIT, $0-48
	MOVQ	a_base+0(FP), DI
	MOVQ	a_len+8(FP), CX
	MOVQ	b_base+24(FP), SI

loop:
	CMPQ	CX, $4
	JB	pair

	// Union four blocks of a with four blocks of b into a.
	MOVOU	( 0*16)(DI), X0
	MOVOU	( 1*16)(DI), X1
	MOVOU	( 2*16)(DI), X2
	MOVOU	( 3*16)(DI), X3
	MOVOU	( 4*16)(DI), X4
	MOVOU	( 5*16)(DI), X5
	MOVOU	( 6*16)(DI), X6
	MOVOU	( 7*16)(DI), X7
	MOVOU	( 8*16)(DI), X8
	MOVOU	( 9*16)(DI), X9
	MOVOU	(10*16)(DI), X10
	MOVOU	(11*16)(DI), X11
	MOVOU	(12*16)(DI), X12
	MOVOU	(13*16)(DI), X13
	MOVOU	(14*16)(DI), X14
	MOVOU	(15*16)(DI), X15

	POR	( 0*16)(SI), X0
	POR	( 1*16)(SI), X1
	POR	( 2*16)(SI), X2
	POR	( 3*16)(SI), X3
	POR	( 4*16)(SI), X4
	POR	( 5*16)(SI), X5
	POR	( 6*16)(SI), X6
	POR	( 7*16)(SI), X7
	POR	( 8*16)(SI), X8
	POR	( 9*16)(SI), X9
	POR	(10*16)(SI), X10
	POR	(11*16)(SI), X11
	POR	(12*16)(SI), X12
	POR	(13*16)(SI), X13
	POR	(14*16)(SI), X14
	POR	(15*16)(SI), X15

	MOVOU	X0,  ( 0*16)(DI)
	MOVOU	X1,  ( 1*16)(DI)
	MOVOU	X2,  ( 2*16)(DI)
	MOVOU	X3,  ( 3*16)(DI)
	MOVOU	X4,  ( 4*16)(DI)
	MOVOU	X5,  ( 5*16)(DI)
	MOVOU	X6,  ( 6*16)(DI)
	MOVOU	X7,  ( 7*16)(DI)
	MOVOU	X8,  ( 8*16)(DI)
	MOVOU	X9,  ( 9*16)(DI)
	MOVOU	X10, (10*16)(DI)
	MOVOU	X11, (11*16)(DI)
	MOVOU	X12, (12*16)(DI)
	MOVOU	X13, (13*16)(DI)
	MOVOU	X14, (14*16)(DI)
	MOVOU	X15, (15*16)(DI)

	SUBQ	$4, CX
	LEAQ	(256)(DI), DI
	LEAQ	(256)(SI), SI
	JMP	loop

pair:
	CMPQ	CX, $2
	JB	single

	MOVOU	(0*16)(DI), X0
	MOVOU	(1*16)(DI), X1
	MOVOU	(2*16)(DI), X2
	MOVOU	(3*16)(DI), X3
	MOVOU	(4*16)(DI), X4
	MOVOU	(5*16)(DI), X5
	MOVOU	(6*16)(DI), X6
	MOVOU	(7*16)(DI), X7

	POR	(0*16)(SI), X0
	POR	(1*16)(SI), X1
	POR	(2*16)(SI), X2
	POR	(3*16)(SI), X3
	POR	(4*16)(SI), X4
	POR	(5*16)(SI), X5
	POR	(6*16)(SI), X6
	POR	(7*16)(SI), X7

	MOVOU	X0, (0*16)(DI)
	MOVOU	X1, (1*16)(DI)
	MOVOU	X2, (2*16)(DI)
	MOVOU	X3, (3*16)(DI)
	MOVOU	X4, (4*16)(DI)
	MOVOU	X5, (5*16)(DI)
	MOVOU	X6, (6*16)(DI)
	MOVOU	X7, (7*16)(DI)

single:
	CMPQ	CX, $0
	JE	end

	MOVOU	(0*16)(DI), X0
	MOVOU	(1*16)(DI), X1
	MOVOU	(2*16)(DI), X2
	MOVOU	(3*16)(DI), X3

	MOVOU	(0*16)(SI), X8
	MOVOU	(1*16)(SI), X9
	MOVOU	(2*16)(SI), X10
	MOVOU	(3*16)(SI), X11

	POR	X0, X8
	POR	X1, X9
	POR	X2, X10
	POR	X3, X11

	MOVOU	X8,  (0*16)(DI)
	MOVOU	X9,  (1*16)(DI)
	MOVOU	X10, (2*16)(DI)
	MOVOU	X11, (3*16)(DI)

end:
	RET
