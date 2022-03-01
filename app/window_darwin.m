// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

@import AppKit;
@import Dispatch;
@import Foundation;
@import QuartzCore;

#include "_cgo_export.h"

void polyred_wakeupMainThread(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		polyred_dispatchMainFuncs();
	});
}

@interface PolyredAppDelegate : NSObject<NSApplicationDelegate>
@end

@interface PolyredWindowDelegate : NSObject<NSWindowDelegate>
@end

@implementation PolyredAppDelegate
- (void)applicationDidFinishLaunching:(NSNotification *)aNotification {
	[NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
	[NSApp activateIgnoringOtherApps:YES];
	polyred_onFinishLaunching();
}
- (void)applicationDidHide:(NSNotification *)aNotification {
	polyred_onAppHide();
}
- (void)applicationWillUnhide:(NSNotification *)notification {
	polyred_onAppShow();
}
@end

@implementation PolyredWindowDelegate
- (void)windowWillMiniaturize:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	polyred_onHide((__bridge CFTypeRef)window.contentView);
}
- (void)windowDidDeminiaturize:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	polyred_onShow((__bridge CFTypeRef)window.contentView);
}
- (void)windowDidChangeScreen:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	CGDirectDisplayID dispID = [[[window screen] deviceDescription][@"NSScreenNumber"] unsignedIntValue];
	CFTypeRef view = (__bridge CFTypeRef)window.contentView;
	polyred_onChangeScreen(view, dispID);
}
- (void)windowDidBecomeKey:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	polyred_onFocus((__bridge CFTypeRef)window.contentView, 1);
}
- (void)windowDidResignKey:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	polyred_onFocus((__bridge CFTypeRef)window.contentView, 0);
}
- (void)windowWillClose:(NSNotification *)notification {
	NSWindow *window = (NSWindow *)[notification object];
	window.delegate = nil;
	polyred_onClose((__bridge CFTypeRef)window.contentView);
}
@end

// Keep a single global reference because delegates are weakly
// referenced from their peers. Nothing else holds a strong
// reference to our window delegate.
static PolyredWindowDelegate *globalWindowDel;

void polyred_main() {
	@autoreleasepool{
		[NSApplication sharedApplication];
		PolyredAppDelegate *del = [[PolyredAppDelegate alloc] init];
		[NSApp setDelegate:del];

		NSMenuItem *mainMenu = [NSMenuItem new];

		NSMenu *menu = [NSMenu new];
		NSMenuItem *hideMenuItem = [[NSMenuItem alloc] initWithTitle:@"Hide"
															  action:@selector(hide:)
													   keyEquivalent:@"h"];
		[menu addItem:hideMenuItem];
		NSMenuItem *quitMenuItem = [[NSMenuItem alloc] initWithTitle:@"Quit"
															  action:@selector(terminate:)
													   keyEquivalent:@"q"];
		[menu addItem:quitMenuItem];
		[mainMenu setSubmenu:menu];
		NSMenu *menuBar = [NSMenu new];
		[menuBar addItem:mainMenu];
		[NSApp setMainMenu:menuBar];

		globalWindowDel = [[PolyredWindowDelegate alloc] init];

		[NSApp run];
	}
}

CFTypeRef polyred_createWindow(
	CFTypeRef viewRef,
	const char *title,
	CGFloat width,
	CGFloat height,
	CGFloat minWidth,
	CGFloat minHeight,
	CGFloat maxWidth,
	CGFloat maxHeight,
	bool canResize
) {
	@autoreleasepool {
		NSRect rect = NSMakeRect(0, 0, width, height);
		NSUInteger styleMask = NSWindowStyleMaskTitled |
			NSWindowStyleMaskMiniaturizable |
			NSWindowStyleMaskClosable;
		if (canResize) {
			styleMask |= NSWindowStyleMaskResizable;
		}

		NSWindow* window = [[NSWindow alloc] initWithContentRect:rect
								styleMask:styleMask
								backing:NSBackingStoreBuffered
								defer:NO];
		if (minWidth > 0 || minHeight > 0) {
			window.contentMinSize = NSMakeSize(minWidth, minHeight);
		}
		if (maxWidth > 0 || maxHeight > 0) {
			window.contentMaxSize = NSMakeSize(maxWidth, maxHeight);
		}
		[window setAcceptsMouseMovedEvents:YES];
		if (title != nil) {
			window.title = [NSString stringWithUTF8String: title];
		}
		NSView *view = (__bridge NSView *)viewRef;
		[window setContentView:view];
		[window makeFirstResponder:view];
		window.releasedWhenClosed = NO;
		window.delegate = globalWindowDel;
		return (__bridge_retained CFTypeRef)window;
	}
}

static void handleMouse(NSView *view, NSEvent *event, int typ, CGFloat dx, CGFloat dy) {
	NSPoint p = [view convertPoint:[event locationInWindow] fromView:nil];
	if (!event.hasPreciseScrollingDeltas) {
		// dx and dy are in rows and columns.
		dx *= 10;
		dy *= 10;
	}
	// Origin is in the lower left corner. Convert to upper left.
	CGFloat height = view.bounds.size.height;
	polyred_onMouse((__bridge CFTypeRef)view, typ,
		[NSEvent pressedMouseButtons], p.x, height - p.y, dx, dy,
		[event timestamp], [event modifierFlags] & NSEventModifierFlagDeviceIndependentFlagsMask);
}

@interface PolyredView : NSView <CALayerDelegate>
@end

@implementation PolyredView

// drawRect is called when OpenGL is used. Don't use OpenGL on darwin.
// - (void)drawRect:(NSRect)r {
// 	polyred_onDraw((__bridge CFTypeRef)self);
// }

// displayLayer is called when Metal is used
- (void)displayLayer:(CALayer *)layer {
	layer.contentsScale = self.window.backingScaleFactor;
	polyred_onDraw((__bridge CFTypeRef)self);
}
- (CALayer *)makeBackingLayer {
	CALayer *layer = [[CAMetalLayer alloc] init];
	layer.delegate = self;
	return layer;
}
- (void)mouseDown:(NSEvent *)event {
	handleMouse(self, event, MOUSE_DOWN, 0, 0);
}
- (void)mouseUp:(NSEvent *)event {
	handleMouse(self, event, MOUSE_UP, 0, 0);
}
- (void)middleMouseDown:(NSEvent *)event {
	handleMouse(self, event, MOUSE_DOWN, 0, 0);
}
- (void)middletMouseUp:(NSEvent *)event {
	handleMouse(self, event, MOUSE_UP, 0, 0);
}
- (void)rightMouseDown:(NSEvent *)event {
	handleMouse(self, event, MOUSE_DOWN, 0, 0);
}
- (void)rightMouseUp:(NSEvent *)event {
	handleMouse(self, event, MOUSE_UP, 0, 0);
}
- (void)mouseMoved:(NSEvent *)event {
	handleMouse(self, event, MOUSE_MOVE, 0, 0);
}
- (void)mouseDragged:(NSEvent *)event {
	handleMouse(self, event, MOUSE_MOVE, 0, 0);
}
- (void)scrollWheel:(NSEvent *)event {
	CGFloat dx = event.scrollingDeltaX;
	CGFloat dy = event.scrollingDeltaY;
	handleMouse(self, event, MOUSE_SCROLL, dx, dy);
}
- (void)keyDown:(NSEvent *)event {
	NSString *keys = [event charactersIgnoringModifiers];
	unsigned short keycode = [event keyCode];
	polyred_onKeys((__bridge CFTypeRef)self, keycode, (char *)[keys UTF8String],
		[event timestamp], [event modifierFlags] & NSEventModifierFlagDeviceIndependentFlagsMask, true);
	[self interpretKeyEvents:[NSArray arrayWithObject:event]];
}
- (void)keyUp:(NSEvent *)event {
	NSString *keys = [event charactersIgnoringModifiers];
	unsigned short keycode = [event keyCode];
	polyred_onKeys((__bridge CFTypeRef)self, keycode, (char *)[keys UTF8String],
		[event timestamp], [event modifierFlags] & NSEventModifierFlagDeviceIndependentFlagsMask, false);
}
- (void)insertText:(id)string {
	const char *utf8 = [string UTF8String];
	polyred_onText((__bridge CFTypeRef)self, (char *)utf8);
}
- (void)doCommandBySelector:(SEL)sel {
	// Don't pass commands up the responder chain.
	// They will end up in a beep.
}
@end

CFTypeRef polyred_createView(void) {
	@autoreleasepool {
		NSRect frame = NSMakeRect(0, 0, 0, 0);
		PolyredView* view = [[PolyredView alloc] initWithFrame:frame];
		[view setWantsLayer:YES];
		view.layerContentsRedrawPolicy = NSViewLayerContentsRedrawDuringViewResize;
		return CFBridgingRetain(view);
	}
}
