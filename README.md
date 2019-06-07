# Message Delivery System

## Table of contents

* [Introduction](#introduction)
* [Protocol](#protocol)
* [Assumptions/Decisions](#assumptionsdecisions)
* [Setup](#setup)
* [Running](#running-the-program)
* [Testing](#testing-the-program)

## Introduction

A _message delivery system_,
including both the server(hub) and the client.

#### Features

1. Identity message - Client can send a identity message which the hub will answer with the user_id of the connected user.
2. List message - Client can send a list message which the hub will answer with the list of all connected client user_id:s (excluding the requesting client).
3. Relay message

## Protocol

 - Protocol is on top of pure TCP.
 - Message types: `who_am_i`, `who_is_here`, `relay`.
 - For request of message types: `who_am_i` and `who_is_here`, the protocol is:
        
        [MessageTypeLength - 1 byte][MessasgeType]
        
  - For response of message type: `who_am_i`, the protocol is:
         
         [userID - 8 bytes]
         
  - For response of message type: `who_is_here`, the protocol is:
          
          [userIDsLength - 1 byte][UserIDs]
        
 - For request of message type: `relay`, the protocol is:
         
        [MessageTypeLength - 1 byte][MessasgeType][ReceiverListLength - 1 byte][Receivers][MessageLength - 4 bytes][Message]
       
 - For response of message type: `relay`, the protocol is:
         
         [userID - 8 bytes][MessageLength - 4 bytes][Message]
        
## Assumptions/Decisions

## Setup

## Running the program

## Testing the program