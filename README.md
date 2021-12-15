# Blockchain For Education

Feel free to send me any feedback that you may have about this project, or alternatively, feel free to make any changes to the repository within a new branch and start a pull request.

Click [here](media/BlockchainArchitecture.pdf) to view the Blockchain Architecture.

To view a presentation on the project, find a PDF version [here](media/ProjectOverview.pdf).

To view a demonstration of the project, find a short video [here](media/demo.mp4)

Contact information is as follows:

Name: Zack Holmberg  
Email: zack_holmberg@outlook.com

## Setting up and running the system

- Prequisites:

  - Go 1.17
  - UNIX environment

- First, clone the repository to your local machine.
- Then, open up at least three terminal windows (one for Middlware, at least two for Peers)
- Navigate one terminal window inside `src/middlware/`
- Navigate at least two other terminal window sinside `src/peer/`
- **NOTE:** If you encounter the error

  ```text
  go: cannot find main module
  ```

  when running the executables, you will have to add the `blockchain` source folder to your `GOROOT` directories. To do this, simply run `go env` in your terminal, navigate inside the directory that was printed, and copy and paste the entire `src/blockchain` directory into your `GOROOT` directory.

- You must run the Middlware executable **FIRST**. To run the Middleware, enter `go run main.go` in the Middleware terminal window

  - Note: For any node that you run, you may get a popup similar to the following:

    ```text
    Do you want the application “main” to accept incoming network connections?
    ```

    Just hit 'allow' if this popup does occur.

- After the Middleware is running, enter the same command, `go run main.go`, in all the Peer terminal windows.
- Each node will take a few seconds to initialize. Once you see messages being logged (Prefixed with date and time), the node is ready for use.

## Using the system

- The system can be interacted with by executing commands through the Peer node. The Middleware has no interaction, but it logs a lot of information about the blockchain's current state.
- To interact with the Peer, you simply enter the command you desire into the terminal window. All current commands are:

| Command       | Description                                                                                                                                                                                    | Example Output                                                                |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `help`        | Lists all valid commands with their descriptions.                                                                                                                                              | The content of this table                                                     |
| `transaction` | Prompts user for recipient and amount values to send a new transaction. Expected input is of the form 'index of recipient in peers list,amount'. For example, '1,5' excluding the apostrophes. | Enter transaction data or 'cancel' to cancel.                                 |
| `peers`       | Lists all of the peers on the network that the user can send currency to.                                                                                                                      | index=0, ip=::1, port=8080 [Middleware Peer]<br />index=1, ip=::1, port=55083 |
| `bal`         | Prints out the user's current wallet balance.                                                                                                                                                  | 10                                                                            |

- For example, after you run a couple Peers, you can enter `peers` in one of the Peers' terminal windows to get a list of known Peers, followed by `transaction` and then `1,5` to send 5 units of currency to the Peer at index 1 of the Peers list. You cannot send currency to the Middleware, only fellow Peers. If you attempt to do so, you will get a warning and no transaction will occur. If you successfully send a transaction to a fellow Peer, a new mining session will occur.

## Swapping Component Implementations

- The system supports the swapping of the Proof of Work and Proof of Stake consensus mechanisms/components to use in the system. A Peer can be created with either implementation, however every Peer on the network must be using the same consensus mechanism.
- To swap between the two components, simply open `src/peer/main.go` and comment-out the Peer initialization with the conensus method you do not want to use. To swap components, comment-out the one implemenation and un-comment the other.
