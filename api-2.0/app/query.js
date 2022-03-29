const { Gateway, Wallets, } = require('fabric-network');
const fs = require('fs');
const path = require("path")
const log4js = require('log4js');
const logger = log4js.getLogger('BasicNetwork');
const util = require('util')
var config = require('./config.json')

const helper = require('./helper')
const query = async (channelName, chaincodeName, args, fcn, username, org_name) => {

    try {

        // load the network configuration
        // const ccpPath = path.resolve(__dirname, '..', 'config', 'connection-org1.json');
        // const ccpJSON = fs.readFileSync(ccpPath, 'utf8')
        const ccp = await helper.getCCP(org_name) //JSON.parse(ccpJSON);

        // Create a new file system based wallet for managing identities.
        const walletPath = await helper.getWalletPath(org_name) //.join(process.cwd(), 'wallet');
        const wallet = await Wallets.newFileSystemWallet(walletPath);
        console.log(`Wallet path: ${walletPath}`);

        // Check to see if we've already enrolled the user.
        let identity = await wallet.get(username);
        if (!identity) {
            console.log(`An identity for the user ${username} does not exist in the wallet, so registering user`);
            await helper.getRegisteredUser(username, org_name, true)
            identity = await wallet.get(username);
            console.log('Run the registerUser.js application before retrying');
            return;
        }

        // Create a new gateway for connecting to our peer node.
        const gateway = new Gateway();
        await gateway.connect(ccp, {
            wallet, identity: username, discovery: { enabled: true, asLocalhost: true }
        });

        // Get the network (channel) our contract is deployed to.
        const network = await gateway.getNetwork(channelName);

        // Get the contract from the network.
        const contract = network.getContract(chaincodeName);
        let result;
        //(fcn == "readPrivateCar" || fcn == "queryPrivateDataHash" || fcn == "collectionCarPrivateDetails")
        if (fcn == 'getHistoryForAsset' || fcn=='restictedMethod' || fcn=='getProfile'  || fcn=='queryRecord' || fcn=='getStatistics' || fcn=='getQuestionnaireByCategory' || fcn=='getAllQuestionnaire' || fcn=='getHistoryForRecord' || fcn=='getReferencesCountByContractorId' || fcn=='getMyReferences' || fcn=='getEmployerProfile' || fcn=='getEmployerStatistics'    || fcn=='getHistoryForAcceptedRequestFromEmployer') {
            result = await contract.evaluateTransaction(fcn, args[0]);
        } else if (fcn == "getAllContractorReferencesByStatus") {  
            result = await contract.evaluateTransaction(fcn, args[0], args[1]);
        } else if (fcn == "getReferencesListByCompanyName") {  
            result = await contract.evaluateTransaction(fcn, args[0], args[1]);
            // return result
        } else if (fcn == "getAllPendingReferences") {  
            result = await contract.evaluateTransaction(fcn);
        } else if (fcn == "getReferencesViewedHistory") {  
            result = await contract.evaluateTransaction(fcn, args[0], args[1]);
        } else if (fcn == "getContractorProfileForEmployer") {  
            result = await contract.evaluateTransaction(fcn, args[0], args[1], args[2]);
        } else if (fcn == "getAllEmployerRequestsByStatus") {  
            result = await contract.evaluateTransaction(fcn, args[0], args[1]);
        }
        
        console.log(result)
        console.log(`Transaction has been evaluated, result is: ${result.toString()}`);

        result = JSON.parse(result.toString());
         const response_payload = {
            status:200,
            result: result,
            error: null,
            errorData: null
        }
        
        return response_payload
       // return result
    } catch (error) {
        console.error(`Failed to evaluate transaction: ${error}`);
       // return error.message
           const response_payload = {
            status:500,
            result: 'fail',
            error: error.message,
            errorData: null
        }
         return response_payload
    }
}

exports.query = query