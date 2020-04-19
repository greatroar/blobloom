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
	CMPQ	CX, $2
	JB	single

	// Union two blocks of a (X0:X7) with two blocks of b (X8:X15) into a (DI).
	MOVOU	(0*16)(DI), X0
	MOVOU	(1*16)(DI), X1
	MOVOU	(2*16)(DI), X2
	MOVOU	(3*16)(DI), X3
	MOVOU	(4*16)(DI), X4
	MOVOU	(5*16)(DI), X5
	MOVOU	(6*16)(DI), X6
	MOVOU	(7*16)(DI), X7

	MOVOU	(0*16)(SI), X8
	MOVOU	(1*16)(SI), X9
	MOVOU	(2*16)(SI), X10
	MOVOU	(3*16)(SI), X11
	MOVOU	(4*16)(SI), X12
	MOVOU	(5*16)(SI), X13
	MOVOU	(6*16)(SI), X14
	MOVOU	(7*16)(SI), X15

	POR	X0, X8
	POR	X1, X9
	POR	X2, X10
	POR	X3, X11
	POR	X4, X12
	POR	X5, X13
	POR	X6, X14
	POR	X7, X15

	MOVOU	X8,  (0*16)(DI)
	MOVOU	X9,  (1*16)(DI)
	MOVOU	X10, (2*16)(DI)
	MOVOU	X11, (3*16)(DI)
	MOVOU	X12, (4*16)(DI)
	MOVOU	X13, (5*16)(DI)
	MOVOU	X14, (6*16)(DI)
	MOVOU	X15, (7*16)(DI)

	SUBQ	$2, CX
	LEAQ	(128)(DI), DI
	LEAQ	(128)(SI), SI
	JMP	loop

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
